from node_instance_watcher import MovingInstances
from dependency_client import etcd_client
from const_var import NS_POS_KEY_TEMPLATE,\
    TYPE_SATELLITE,\
    POLAR_REGION_LATITUDE,\
    PARAMETER_KEY_CONNECT,\
    PARAMETER_KEY_DELAY,\
    NODE_LINK_PARAMETER_KEY_TEMPLATE
from concurrent.futures import ThreadPoolExecutor,wait,ALL_COMPLETED
from datetime import datetime
import time, json
from node_instance_watcher import MovingInstances,MovingInstancesLock
from instance import Instance,distance,get_propagation_delay
from satellite import Satellite
from link import ISL,GSL


def update_etcd(key:str, value:str):
    etcd_client.put(key,value)

def calculate():
    while True:
        now = datetime.now()
        MovingInstancesLock.acquire()
        keys = MovingInstances.keys()
        for id in keys:
            if isinstance(MovingInstances[id],Satellite):
                MovingInstances[id].calculate_postion(now)
        ns_type_position_map : dict[str,dict[str,list[dict]]]= {}
        node_link_paramter_map = {}
        for id in keys:
            instance = MovingInstances[id]
            for link_id,inst_link in instance.links.items():
                # Check Link Parameter
                if isinstance(inst_link,ISL):
                    connected_instances:list[Instance] = [
                        MovingInstances[inst_link.instance_id[0]],
                        MovingInstances[inst_link.instance_id[1]]
                    ]
                    
                    if connected_instances[0].node_index not in node_link_paramter_map:
                        node_link_paramter_map[connected_instances[0].node_index] = {}
                    if connected_instances[1].node_index not in node_link_paramter_map:
                        node_link_paramter_map[connected_instances[1].node_index] = {}

                    if link_id in node_link_paramter_map[connected_instances[0].node_index] and \
                        link_id in node_link_paramter_map[connected_instances[1].node_index]:
                        continue
                    
                    polar_region_link_close = abs(connected_instances[0].latitude) > POLAR_REGION_LATITUDE and\
                        inst_link.is_inter_orbit

                    polar_region_link_close = polar_region_link_close or \
                        (abs(connected_instances[1].latitude) > POLAR_REGION_LATITUDE and inst_link.is_inter_orbit)
                    
                    if polar_region_link_close:
                        inst_link.parameters[PARAMETER_KEY_CONNECT] = 0
                    else:
                        inst_link.parameters[PARAMETER_KEY_CONNECT] = 1
                    inst_link.parameters[PARAMETER_KEY_DELAY] = int(get_propagation_delay(
                        distance(
                            MovingInstances[inst_link.instance_id[0]],
                            MovingInstances[inst_link.instance_id[1]]
                        )
                    ) * 1e6)

                    node_link_paramter_map[connected_instances[0].node_index][link_id] = inst_link.parameters
                    node_link_paramter_map[connected_instances[1].node_index][link_id] = inst_link.parameters
                else:
                    pass

            # Build Position Data Structure
            if instance.namespace not in ns_type_position_map:
                ns_type_position_map[instance.namespace] = {instance.type:{instance.instance_id:instance.get_position_dict()}}
            elif instance.type not in ns_type_position_map[instance.namespace]:
                ns_type_position_map[instance.namespace][instance.type] = {instance.instance_id:instance.get_position_dict()}
            else:
                ns_type_position_map[instance.namespace][instance.type][instance.instance_id] = instance.get_position_dict()

        thread_pool = ThreadPoolExecutor(max_workers=64)
        all_tasks = []

        # Update Position Data
        for ns in ns_type_position_map.keys():
            for instance_type in ns_type_position_map[ns].keys():
                instance_position_key = NS_POS_KEY_TEMPLATE%(ns,instance_type)
                obj_seq = json.dumps(ns_type_position_map[instance.namespace][instance.type])
            print("Update ",obj_seq)
            all_tasks.append(thread_pool.submit(update_etcd,instance_position_key,obj_seq))

        # Update Parameter Data
        for node_index,parameters in node_link_paramter_map.items():
            link_parameter_key = NODE_LINK_PARAMETER_KEY_TEMPLATE%node_index
            obj_seq = json.dumps(parameters)
            all_tasks.append(thread_pool.submit(update_etcd,link_parameter_key,obj_seq))
        

        wait(all_tasks, timeout=None, return_when=ALL_COMPLETED)
        thread_pool.shutdown()
        now = datetime.now()
        MovingInstancesLock.release()
        time.sleep(1)

