from dependency_client import etcd_client,redis_client
from const_var import NS_POS_KEY_TEMPLATE,\
    TYPE_SATELLITE,\
    POLAR_REGION_LATITUDE,\
    PARAMETER_KEY_CONNECT,\
    PARAMETER_KEY_DELAY,\
    NODE_LINK_PARAMETER_KEY_TEMPLATE,\
    NODE_LINK_INFO_KEY_TEMPLATE,\
    NODE_LINK_LIST_KEY_TEMPLATE,\
    LINK_V4_ADDR_KEY
from concurrent.futures import ThreadPoolExecutor,wait,ALL_COMPLETED
from datetime import datetime
from ground_station import GroundStation
import time, json
from instance import Instance,distance,get_propagation_delay,Instances,InstanceLock
from satellite import Satellite
from link import ISL,GSL,LinksLock,LinkConfig,LinkEndInfo,Links
from loguru import logger
from address_allocator import alloc_ipv4,format_ipv4
from tools import object2Map
import uuid

def update_etcd(key:str, value:str):
    etcd_client.put(key,value)

def calculate_satellite_position(now: datetime):
    for id in Instances.keys():
        if isinstance(Instances[id],Satellite):
            Instances[id].calculate_postion(now)

# Return: node_index -> link_id -> parameter_name -> parameter_value
def calculate_ISL_parameters() -> dict[int,dict[str,dict[str,int]]]:
    node_link_paramter_map = {}
    for instance_id in Instances.keys():
        instance = Instances[instance_id]
        if not isinstance(instance,Satellite):
            continue
        
        for link_id,inst_link in instance.links.items():
            
            if isinstance(inst_link,ISL):
                if inst_link.config.init_end_infos[0].instance_id not in Instances.keys() :
                    continue
                if inst_link.config.init_end_infos[1].instance_id not in Instances.keys() :
                    continue
                connected_instances:list[Instance] = [
                    Instances[inst_link.config.init_end_infos[0].instance_id],
                    Instances[inst_link.config.init_end_infos[1].instance_id]
                ]
                
                if connected_instances[0].node_index not in node_link_paramter_map:
                    node_link_paramter_map[connected_instances[0].node_index] = {}
                if connected_instances[1].node_index not in node_link_paramter_map:
                    node_link_paramter_map[connected_instances[1].node_index] = {}
                if link_id in node_link_paramter_map[connected_instances[0].node_index] and \
                    link_id in node_link_paramter_map[connected_instances[1].node_index]:
                    continue
                is_inter_orbit = connected_instances[1].orbit_index != connected_instances[0].orbit_index
                polar_region_link_close = abs(connected_instances[0].latitude) > POLAR_REGION_LATITUDE and\
                    inst_link.is_inter_orbit
                polar_region_link_close = polar_region_link_close or \
                    (abs(connected_instances[1].latitude) > POLAR_REGION_LATITUDE and is_inter_orbit)

                
                if polar_region_link_close:
                    logger.info("disconnect link %s"%link_id,abs(connected_instances[1].latitude) > POLAR_REGION_LATITUDE,is_inter_orbit)
                    inst_link.parameter[PARAMETER_KEY_CONNECT] = 0
                else:
                    inst_link.parameter[PARAMETER_KEY_CONNECT] = 1
                inst_link.parameter[PARAMETER_KEY_DELAY] = int(get_propagation_delay(
                    distance(
                        Instances[inst_link.config.init_end_infos[0].instance_id],
                        Instances[inst_link.config.init_end_infos[1].instance_id]
                    )
                ) * 1e6)
                node_link_paramter_map[connected_instances[0].node_index][link_id] = inst_link.parameter
                node_link_paramter_map[connected_instances[1].node_index][link_id] = inst_link.parameter
    return node_link_paramter_map

# Return namespace -> instance_type -> instance_id -> position_dict
def build_position_data_structure() -> dict[str,dict[str,list[dict]]]:
    keys = Instances.keys()
    ns_type_position_map : dict[str,dict[str,dict[dict[str,float]]]]= {}
    for id in keys:
        instance = Instances[id]
        # Build Position Data Structure
        if instance.namespace not in ns_type_position_map:
            ns_type_position_map[instance.namespace] = {instance.type:{instance.instance_id:instance.get_position_dict()}}
        elif instance.type not in ns_type_position_map[instance.namespace]:
            ns_type_position_map[instance.namespace][instance.type] = {instance.instance_id:instance.get_position_dict()}
        else:
            ns_type_position_map[instance.namespace][instance.type][instance.instance_id] = instance.get_position_dict()
    return ns_type_position_map

# Return Add GSL List, Del GSL List
def judge_ground_station_connection() -> (list[GSL],list[GSL]):
    gsl_add_list: list[GSL] = []
    gsl_del_list: list[GSL] = []
    for (instance_id,instance_info) in Instances.items():
        if not isinstance(instance_info,GroundStation):
            continue
        next_satellite,new_distance = instance_info.get_closest_satellite(Instances)
        if next_satellite == None:
            continue
        if instance_info.connected_satellite_id is None:
            link = GSL({})
            link.enabled = False
            link.parameter = {
                PARAMETER_KEY_CONNECT:1,
                PARAMETER_KEY_DELAY: int(get_propagation_delay(new_distance))
            }
            link.cross_machine = next_satellite.node_index != instance_info.node_index
            link.config = LinkConfig({})
            link.config.link_id = uuid.uuid4().hex[0:8]
            link.config.init_parameter = link.parameter
            link.config.init_end_infos = [LinkEndInfo({}),LinkEndInfo({})]
            link.config.init_end_infos[0].instance_id = instance_id
            link.config.init_end_infos[0].instance_type = instance_info.type
            link.config.init_end_infos[0].node_index = instance_info.node_index
            link.config.init_end_infos[1].instance_id = next_satellite.instance_id
            link.config.init_end_infos[1].instance_type = next_satellite.type
            link.config.init_end_infos[1].node_index = next_satellite.node_index
            link.config.type = "vlink"
            subnet = alloc_ipv4(30)
            link.config.address_infos = [
                {
                    LINK_V4_ADDR_KEY:format_ipv4(subnet[1],30)
                },
                {
                    LINK_V4_ADDR_KEY:format_ipv4(subnet[2],30)
                }
            ]
            instance_info.connect_link_id = link.config.link_id
            instance_info.connected_satellite_id = next_satellite.instance_id
            gsl_add_list.append(link)
            instance_info.links[link.config.link_id] = link
            Links[link.config.link_id] = link
        elif next_satellite.instance_id != instance_info.connected_satellite_id:
            link = GSL({})
            link.enabled = False
            
            link.parameter = {
                PARAMETER_KEY_CONNECT:1,
                PARAMETER_KEY_DELAY: int(get_propagation_delay(new_distance))
            }
            link.cross_machine = next_satellite.node_index != instance_info.node_index
            link.config = LinkConfig({})
            link.config.link_id = uuid.uuid4().hex[0:8]
            link.config.init_parameter = link.parameter
            link.config.init_end_infos = [LinkEndInfo({}),LinkEndInfo({})]
            link.config.init_end_infos[0].instance_id = instance_id
            link.config.init_end_infos[0].instance_type = instance_info.type
            link.config.init_end_infos[0].node_index = instance_info.node_index
            link.config.init_end_infos[1].instance_id = next_satellite.instance_id
            link.config.init_end_infos[1].instance_type = next_satellite.type
            link.config.init_end_infos[1].node_index = next_satellite.node_index
            link.config.type = "vlink"
            subnet = alloc_ipv4(30)
            link.config.address_infos = [
                {
                    LINK_V4_ADDR_KEY:format_ipv4(subnet[1],30)
                },
                {
                    LINK_V4_ADDR_KEY:format_ipv4(subnet[2],30)
                }
            ]
            gsl_del_list.append(instance_info.links[instance_info.connect_link_id])
            instance_info.connect_link_id = link.config.link_id
            instance_info.connected_satellite_id = next_satellite.instance_id
            gsl_add_list.append(link)
            instance_info.connect_link_id = link.config.link_id
            instance_info.connected_satellite_id = next_satellite.instance_id
            instance_info.links[link.config.link_id] = link
            Links[link.config.link_id] = link
    return gsl_add_list,gsl_del_list
    
def do_reconnect(gsl_add_list: list[GSL],gsl_del_list: list[GSL]):
    gsl_del_id_list_map : dict[int,list[str]] = {}
    gsl_add_map : dict[int,dict[str,str]] = {}
    
    for item in gsl_del_list:
        gs = Instances[item.config.init_end_infos[0].instance_id]
        sat = Instances[item.config.init_end_infos[1].instance_id]
        del gs.links[item.config.link_id]
        del sat.links[item.config.link_id]
        gs.connected_satellite_id = None
        if gs.node_index in gsl_del_id_list_map:
            gsl_del_id_list_map[gs.node_index].append(item.config.link_id)
        else:
            gsl_del_id_list_map[gs.node_index] = [item.config.link_id]
        if gs.node_index != sat.node_index :
            if sat.node_index in gsl_del_id_list_map:
                gsl_del_id_list_map[sat.node_index].append(item.config.link_id)
            else:
                gsl_del_id_list_map[sat.node_index] = [item.config.link_id]

    for item in gsl_add_list:
        gs = Instances[item.config.init_end_infos[0].instance_id]
        sat = Instances[item.config.init_end_infos[1].instance_id]
        gs.links[item.config.link_id] = item
        sat.links[item.config.link_id] = item
        gs.connected_satellite_id = sat.instance_id
        if gs.node_index in gsl_add_map:
            gsl_add_map[gs.node_index][item.config.link_id] = json.dumps(object2Map(item))
        else:
            gsl_add_map[gs.node_index] = {item.config.link_id: json.dumps(object2Map(item))}
        if gs.node_index != sat.node_index :
            if sat.node_index in gsl_add_map:
                gsl_add_map[sat.node_index][item.config.link_id] = json.dumps(object2Map(item))
            else:
                gsl_add_map[sat.node_index] = {item.config.link_id: json.dumps(object2Map(item))}
    return gsl_del_id_list_map,gsl_add_map
    

def update_link_data(
    node_link_paramter_map: dict[int,dict[str,dict[str,int]]],
    ns_type_position_map: dict[str,dict[str,list[dict]]],
    gsl_del_id_list_map: dict[int,list[str]],
    gsl_add_map: dict[int,dict[str,str]]
):

    thread_pool = ThreadPoolExecutor(max_workers=64)
    all_tasks = []
    # Update Position Data
    for ns in ns_type_position_map.keys():
        for instance_type in ns_type_position_map[ns].keys():
            instance_position_key = NS_POS_KEY_TEMPLATE%(ns,instance_type)
            obj_seq = json.dumps(ns_type_position_map[ns][instance_type])
            all_tasks.append(thread_pool.submit(update_etcd,instance_position_key,obj_seq))

    # Update Parameter Data
    for node_index,parameters in node_link_paramter_map.items():
        link_parameter_key = NODE_LINK_PARAMETER_KEY_TEMPLATE%node_index
        obj_seq = json.dumps(parameters)
        all_tasks.append(thread_pool.submit(update_etcd,link_parameter_key,obj_seq))
    
    # Update GSL Change
    for node_index, del_id_list in gsl_del_id_list_map.items():
        node_link_list_seq,meta = etcd_client.get(NODE_LINK_LIST_KEY_TEMPLATE%node_index)
        node_link_list = json.loads(node_link_list_seq)
        for del_link_id in del_id_list:
            node_link_list.remove(del_link_id)
        redis_client.hdel(
            NODE_LINK_INFO_KEY_TEMPLATE%node_index,
            *del_id_list
        )
        if node_index in gsl_add_map.keys():
            gsl_add_items = gsl_add_map[node_index]
            for new_link_id in gsl_add_items.keys():
                node_link_list.append(new_link_id)
            redis_client.hmset(
                NODE_LINK_INFO_KEY_TEMPLATE%node_index,
                gsl_add_items
            )
        etcd_client.put(NODE_LINK_LIST_KEY_TEMPLATE%node_index,json.dumps(node_link_list))
        del gsl_add_map[node_index]

    for node_index, gsl_add_items in gsl_add_map.items():
        node_link_list_seq,meta = etcd_client.get(NODE_LINK_LIST_KEY_TEMPLATE%node_index)
        node_link_list = json.loads(node_link_list_seq)
        for new_link_id in gsl_add_items.keys():
            node_link_list.append(new_link_id)
            
        redis_client.hmset(
            NODE_LINK_INFO_KEY_TEMPLATE%node_index,
            gsl_add_items
        )
        etcd_client.put(NODE_LINK_LIST_KEY_TEMPLATE%node_index,json.dumps(node_link_list))

    wait(all_tasks, timeout=None, return_when=ALL_COMPLETED)
    thread_pool.shutdown()

def evnet_generator():
    while True:
        now = datetime.now()
        InstanceLock.acquire()
        LinksLock.acquire()
        # try:
        calculate_satellite_position(now)
        node_link_paramter_map = calculate_ISL_parameters()
        ns_type_position_map = build_position_data_structure()
        gsl_add_list,gsl_del_list = judge_ground_station_connection()
        gsl_del_id_list_map, gsl_add_map= do_reconnect(gsl_add_list,gsl_del_list)
        # except Exception as e:
        #     logger.error(e)
        LinksLock.release()
        InstanceLock.release()
        update_link_data(
            node_link_paramter_map,
            ns_type_position_map,
            gsl_del_id_list_map,
            gsl_add_map
        )
        time.sleep(1)

