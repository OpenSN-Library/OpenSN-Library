from node_instance_watcher import Satellites
from dependency_client import etcd_client,redis_client
from const_var import NS_POS_KEY_TEMPLATE,\
    TYPE_SATELLITE,\
    POLAR_REGION_LATITUDE,\
    PARAMETER_KEY_CONNECT,\
    PARAMETER_KEY_DELAY,\
    NODE_LINK_PARAMETER_KEY_TEMPLATE,\
    NODE_LINK_INFO_KEY_TEMPLATE,\
    NODE_LINK_LIST_KEY_TEMPLATE
from concurrent.futures import ThreadPoolExecutor,wait,ALL_COMPLETED
from datetime import datetime
from ground_station import GroundStation
import time, json
from node_instance_watcher import Satellites,SatellitesLock,GroundStations,GroundStationsLock
from instance import Instance,distance,get_propagation_delay
from satellite import Satellite
from link import ISL,GSL
import math


def update_etcd(key:str, value:str):
    etcd_client.put(key,value)

def calculate_satellite_position(now: datetime):
    keys = Satellites.keys()
    for id in keys:
        Satellites[id].calculate_postion(now)

# Return: node_index -> link_id -> parameter_name -> parameter_value
def calculate_ISL_parameters() -> dict[int,dict[str,dict[str,int]]]:
    keys = Satellites.keys()
    node_link_paramter_map = {}
    for id in keys:
        instance = Satellites[id]
        for link_id,inst_link in instance.links.items():
            if isinstance(inst_link,ISL):
                connected_instances:list[Instance] = [
                    Satellites[inst_link.instance_id[0]],
                    Satellites[inst_link.instance_id[1]]
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
                        Satellites[inst_link.instance_id[0]],
                        Satellites[inst_link.instance_id[1]]
                    )
                ) * 1e6)
                node_link_paramter_map[connected_instances[0].node_index][link_id] = inst_link.parameters
                node_link_paramter_map[connected_instances[1].node_index][link_id] = inst_link.parameters
    return node_link_paramter_map

# Return namespace -> instance_type -> instance_id -> position_dict
def build_position_data_structure() -> dict[str,dict[str,list[dict]]]:
    keys = Satellites.keys()
    ns_type_position_map : dict[str,dict[str,dict[dict[str,float]]]]= {}
    for id in keys:
        instance = Satellites[id]
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
    for (gs_id,ground_station) in GroundStations.items():
        next_satellite,new_distance = ground_station.get_closest_satellite(Satellites)
        if ground_station.connected_satellite_id is None:
            gsl_add_list.append(GSL(
                [gs_id,next_satellite.instance_id],
                {
                    PARAMETER_KEY_DELAY:int(get_propagation_delay(new_distance))
                }
            ))
        elif next_satellite.instance_id != ground_station.connected_satellite_id:
            gsl_del_list.append(Satellites[ground_station.connected_satellite_id])
            gsl_add_list.append(GSL(
                [gs_id,next_satellite.instance_id],
                {
                    PARAMETER_KEY_DELAY:int(get_propagation_delay(new_distance))
                }
            ))
    return gsl_add_list,gsl_del_list
    
def do_reconnect(gsl_add_list: list[GSL],gsl_del_list: list[GSL]):
    gsl_del_id_list_map : dict[int,list[str]] = {}
    gsl_add_map : dict[int,dict[str,str]] = {}
    
    for item in gsl_del_list:
        gs = GroundStations[item.instance_id[0]]
        sat = Satellites[item.instance_id[1]]
        del gs.links[item.link_id]
        del sat.links[item.link_id]
        gs.connected_satellite_id = None
        if gs.node_index in gsl_del_id_list_map:
            gsl_del_id_list_map[gs.node_index].append(item.link_id)
        else:
            gsl_del_id_list_map[gs.node_index] = [item.link_id]
        if gs.node_index != sat.node_index :
            if sat.node_index in gsl_del_id_list_map:
                gsl_del_id_list_map[sat.node_index].append(item.link_id)
            else:
                gsl_del_id_list_map[sat.node_index] = [item.link_id]

    for item in gsl_add_list:
        gs = GroundStations[item.instance_id[0]]
        sat = Satellites[item.instance_id[1]]
        gs.links[item.link_id] = item
        sat.links[item.link_id] = item
        gs.connected_satellite_id = sat.instance_id
        if gs.node_index in gsl_add_map:
            gsl_add_map[gs.node_index][item.link_id] = json.dumps(item)
        else:
            gsl_add_map[gs.node_index] = {item.link_id: json.dumps(item)}
        if gs.node_index != sat.node_index :
            if sat.node_index in gsl_add_map:
                gsl_add_map[sat.node_index][item.link_id] = json.dumps(item)
            else:
                gsl_add_map[sat.node_index] = {item.link_id: json.dumps(item)}
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
            print("Update ",instance_position_key)
            all_tasks.append(thread_pool.submit(update_etcd,instance_position_key,obj_seq))

    # Update Parameter Data
    for node_index,parameters in node_link_paramter_map.items():
        link_parameter_key = NODE_LINK_PARAMETER_KEY_TEMPLATE%node_index
        obj_seq = json.dumps(parameters)
        all_tasks.append(thread_pool.submit(update_etcd,link_parameter_key,obj_seq))
    
    # Update GSL Change
    for node_index, del_id_list in gsl_del_id_list_map.items():
        redis_client.hdel(
            NODE_LINK_INFO_KEY_TEMPLATE%node_index,
            del_id_list
        )

    for node_index, gsl_add_items in gsl_add_map.items():
        redis_client.hmset(
            NODE_LINK_INFO_KEY_TEMPLATE%node_index,
            gsl_add_items
        )

    wait(all_tasks, timeout=None, return_when=ALL_COMPLETED)
    thread_pool.shutdown()

def evnet_generator():
    while True:
        now = datetime.now()
        SatellitesLock.acquire()
        GroundStationsLock.acquire()
        calculate_satellite_position(now)
        node_link_paramter_map = calculate_ISL_parameters()
        ns_type_position_map = build_position_data_structure()
        gsl_add_list,gsl_del_list = judge_ground_station_connection()
        gsl_del_id_list_map, gsl_add_map= do_reconnect(gsl_add_list,gsl_del_list)
        GroundStationsLock.release()
        SatellitesLock.release()
        update_link_data(
            node_link_paramter_map,
            ns_type_position_map,
            gsl_del_id_list_map,
            gsl_add_map
        )
        time.sleep(1)

