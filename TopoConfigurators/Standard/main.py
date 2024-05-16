from opensn.operator.emulator_operator import EmulatorOperator
from opensn.model.instance import Instance
from opensn.model.position import Position
from opensn.const.dict_fields import PARAMETER_KEY_CONNECT,PARAMETER_KEY_DELAY,PARAMETER_KEY_BANDWIDTH,PARAMETER_KEY_LOSS
from opensn.model.link import LinkBase
from opensn.utils.tools import dec2ra
from config import ADDR,PORT
from datetime import datetime
from trajectory import calculate_postion,distance_meter,select_closest_satellite,get_propagation_delay_s
from instance_types import TYPE_GROUND_STATION, TYPE_SATELLITE, EX_ORBIT_INDEX,EX_ALTITUDE_KEY,EX_LATITUDE_KEY,EX_LONGITUDE_KEY, EX_AREA_KEY
from address_type import LINK_V4_ADDR_KEY
from time import sleep
from address_allocator import alloc_ipv4,format_ipv4
from loguru import logger
import json, math
step_second = 5

polar_threshold = dec2ra(66.5)

def genenrate_config(cli:EmulatorOperator,node_index:int,instance_id:str):
    instance_info = cli.get_instance(node_index,instance_id)
    config_map = {
        "instance_id": instance_id,
        "link_infos": {},
        "end_infos": {},
    }
    if instance_info.type == TYPE_SATELLITE:
        config_map['area'] = instance_info.extra[EX_AREA_KEY]
    for k,v in instance_info.connections.items():
        instance_index = -1
        link_info = cli.get_link(node_index,k)
        for end_index in range(len(link_info.end_infos)):
            if link_info.end_infos[end_index].instance_id == instance_id:
                instance_index = end_index
        if instance_index < 0:
            return {}
        another_instance_info = cli.get_instance(link_info.end_infos[1-instance_index].end_node_index,link_info.end_infos[1-instance_index].instance_id)
        config_map["link_infos"][k] = link_info.address_infos[instance_index]
        config_map["end_infos"][k] = {
            "instance_id": v.instance_id,
            "type": v.instance_type,
        }
        if another_instance_info.type == TYPE_SATELLITE:
            config_map["end_infos"][k]['area'] = another_instance_info.extra[EX_AREA_KEY]
    return config_map

if __name__ == "__main__":

    instance_config_updated:dict[str,str] = {}
    
    cli = EmulatorOperator(ADDR,PORT)

    # Create Emulator Operator
    while True:
        node_list = cli.get_node_map()
        all_instance_map: dict[str,Instance] = {}
        node_link_map: dict[int,dict[str,LinkBase]] = {}
        ground_station_list:list[Instance] = []
        for node_index,node in node_list.items():
            instance_map = cli.get_instance_map(node_index)
            for instance_id,instance in instance_map.items():
                all_instance_map[instance_id] = instance
                if instance.type == TYPE_GROUND_STATION:
                    ground_station_list.append(instance)
                    gs_position = Position()
                    gs_position.latitude = float(instance.extra[EX_LATITUDE_KEY]) / 180 * math.pi
                    gs_position.longitude = float(instance.extra[EX_LONGITUDE_KEY]) / 180 * math.pi
                    gs_position.altitude = float(instance.extra[EX_ALTITUDE_KEY])
                    cli.put_position(instance_id,gs_position)

        address_map = {}
        for node_index,node in node_list.items():
            node_link_map[node_index] = {}
            link_map = cli.get_link_map(node_index)
            for link_id,link_info in link_map.items():
                if link_info.address_infos[0] is None or \
                    link_info.address_infos[1] is None:
                    if link_id not in address_map.keys():
                        address_map[link_id] = alloc_ipv4(30)
                    
                    subnet = address_map[link_id]
                    link_info.address_infos = [{
                        LINK_V4_ADDR_KEY: format_ipv4(subnet[1],30)
                    },
                    {
                        LINK_V4_ADDR_KEY: format_ipv4(subnet[2],30)
                    }]
                    cli.put_link(link_info)
                node_link_map[node_index][link_id] = link_info
                

        position_map: dict[str,Position] = {"":Position()}
        time_now = datetime.now()
        for instance_id,instance_info in all_instance_map.items():
            if instance_info.start:
                new_postion = calculate_postion(instance_info,time_now)
                cli.put_position(instance_id,new_postion)
            else:
                new_postion = Position()
            position_map[instance_id] = new_postion
        # Do Ground Station Reconnect
    

        for ground_station in ground_station_list:
            if not ground_station.start:
                continue
            gs_position = position_map[ground_station.instance_id]
            satellite_id,change = select_closest_satellite(
                ground_station,
                position_map,
                all_instance_map
            )
            if change:
                address1 = {}
                address2 = {}
                
                old_link_id = ""
                for key in ground_station.connections.keys():
                    old_link_id = key
                    break
                if old_link_id != "":
                    old_link = cli.disable_link_between(
                        ground_station.node_index,
                        ground_station.instance_id,
                        ground_station.connections[key].end_node_index,
                        ground_station.connections[key].instance_id
                    )
                    logger.info("Switch %s from %s to %s"%(
                        ground_station.instance_id,
                        ground_station.connections[old_link_id].instance_id,
                        satellite_id
                    ))
                    old_sat_config = genenrate_config(cli,ground_station.connections[old_link_id].end_node_index,ground_station.connections[old_link_id].instance_id)
                    cli.put_instance_config(ground_station.connections[old_link_id].end_node_index,ground_station.connections[old_link_id].instance_id,json.dumps(old_sat_config))

                    # config_map = genenrate_config(
                    #     satellite_id,all_instance_map[ground_station.connections[key].instance_id],node_link_map)
                    if len(old_link) > 0:
                        address1 = old_link[old_link_id].address_infos[0]
                        address2 = old_link[old_link_id].address_infos[1]
                else:
                    subnet = alloc_ipv4(30)
                    address1 = {LINK_V4_ADDR_KEY:format_ipv4(subnet[1],30)}
                    address2 = {LINK_V4_ADDR_KEY:format_ipv4(subnet[2],30)}
                    logger.info("Switch %s from %s to %s"%(
                        ground_station.instance_id,
                        "None",
                        satellite_id
                    ))
                cli.enable_link_between(
                    ground_station.node_index,
                    ground_station.instance_id,
                    all_instance_map[satellite_id].node_index,
                    all_instance_map[satellite_id].instance_id,
                    address_info1=address1,
                    address_info2=address2,
                )
                gs_config = genenrate_config(cli,ground_station.node_index,ground_station.instance_id)
                # print(gs_config)
                cli.put_instance_config(ground_station.node_index,ground_station.instance_id,json.dumps(gs_config))
                sat_config = genenrate_config(cli,all_instance_map[satellite_id].node_index,all_instance_map[satellite_id].instance_id)
                # print(sat_config)
                cli.put_instance_config(all_instance_map[satellite_id].node_index,all_instance_map[satellite_id].instance_id,json.dumps(sat_config))

        

        for node_index,link_map in node_link_map.items():
            for link_id,link_info in link_map.items():
                if link_info.parameter is None:
                    link_info.parameter = {}
                if link_info.end_infos[0].instance_id=="" or link_info.end_infos[1].instance_id == "":
                    continue
                if not link_info.enable:
                    continue
                if link_info.end_infos[0].instance_type == TYPE_SATELLITE and \
                    link_info.end_infos[1].instance_type == TYPE_SATELLITE and \
                    all_instance_map[link_info.end_infos[1].instance_id].extra[EX_ORBIT_INDEX] != \
                    all_instance_map[link_info.end_infos[0].instance_id].extra[EX_ORBIT_INDEX] and \
                    (abs(position_map[link_info.end_infos[0].instance_id].latitude) > polar_threshold or \
                    abs(position_map[link_info.end_infos[1].instance_id].latitude) > polar_threshold):
                        # if PARAMETER_KEY_CONNECT in link_info.parameter.keys() and link_info.parameter[PARAMETER_KEY_CONNECT]==1:
                        #     logger.info("connect %s"%link_id)
                        link_info.parameter[PARAMETER_KEY_CONNECT] = 0
                else:
                    # if PARAMETER_KEY_CONNECT not in link_info.parameter.keys() or link_info.parameter[PARAMETER_KEY_CONNECT]==0:
                    #         logger.info("disconnect %s"%link_id)
                    link_info.parameter[PARAMETER_KEY_CONNECT] = 1

                distance = distance_meter(
                    position_map[link_info.end_infos[0].instance_id],
                    position_map[link_info.end_infos[1].instance_id]
                )
                delay = int(get_propagation_delay_s(distance)*1000000)
                link_info.parameter[PARAMETER_KEY_DELAY] = delay
                link_info.parameter[PARAMETER_KEY_BANDWIDTH] = 1000000
                link_info.parameter[PARAMETER_KEY_LOSS] = 150
                cli.put_link_parameter(link_info.node_index,link_info.link_id,link_info.parameter)
                
        for instance_id,instance_info in all_instance_map.items():
            if not instance_info.start:
                continue
            config_map = genenrate_config(cli,instance_info.node_index,instance_id)
            cli.put_instance_config_if_not_exist(instance_info.node_index,instance_id,json.dumps(config_map))
        sleep(step_second)
