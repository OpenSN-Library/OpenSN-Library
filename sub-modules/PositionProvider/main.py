from satellite_emulator.operator.emulator_operator import EmulatorOperator
from satellite_emulator.model.instance import Instance
from satellite_emulator.model.position import Position
from satellite_emulator.const.dict_fields import PARAMETER_KEY_CONNECT,PARAMETER_KEY_DELAY
from satellite_emulator.model.link import LinkBase
from config import ADDR,PORT
from datetime import datetime
from trajectory import calculate_postion,distance_meter,select_closest_satellite,get_propagation_delay_s
from instance_types import TYPE_GROUND_STATION, TYPE_SATELLITE, EX_ORBIT_INDEX
from time import sleep
from satellite_emulator.utils.tools import dec2ra
from loguru import logger

step_second = 5

polar_threshold = dec2ra(66.5)

if __name__ == "__main__":
    cli = EmulatorOperator(ADDR,PORT)

    # Create Emulator Operator
    while True:
        node_list = cli.get_node_map()
        all_instance_map: dict[str,Instance] = {}
        ground_station_list:list[Instance] = []
        for node_index,node in node_list.items():
            instance_map = cli.get_instance_map(node_index)
            for instance_id,instance in instance_map.items():
                if instance.start:
                    all_instance_map[instance_id] = instance
                    if instance.type == TYPE_GROUND_STATION:
                        ground_station_list.append(instance)

        position_map: dict[str,Position] = {}
        time_now = datetime.now()
        for instance_id,instance_info in all_instance_map.items():
            new_postion = calculate_postion(instance_info,time_now)
            position_map[instance_id] = new_postion
        # Do Ground Station Reconnect
        
        logger.info(ground_station_list)

        for ground_station in ground_station_list:
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
                    if len(old_link) > 0:
                        address1 = old_link[old_link_id].address_infos[0]
                        address2 = old_link[old_link_id].address_infos[1]
                else:
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

        node_link_map: dict[int,dict[str,LinkBase]] = {}
        
        for node_index,node in node_list.items():
            node_link_map[node_index] = {}
            link_map = cli.get_link_map(node_index)
            for link_id,link_info in link_map.items():
                if link_info.enabled:
                    node_link_map[node_index][link_id] = link_info

        for node_index,link_map in node_link_map.items():
            for link_id,link_info in link_map.items():
                logger.info(link_info)
                if link_info.end_infos[0].instance_type == TYPE_SATELLITE and \
                    link_info.end_infos[1].instance_type == TYPE_SATELLITE and \
                    all_instance_map[link_info.end_infos[1].instance_id].extra[EX_ORBIT_INDEX] == \
                    all_instance_map[link_info.end_infos[0].instance_id].extra[EX_ORBIT_INDEX]:
                    if abs(position_map[link_info.end_infos[0].instance_id].latitude) > polar_threshold or \
                        abs(position_map[link_info.end_infos[1].instance_id].latitude) > polar_threshold:
                        link_info.parameter[PARAMETER_KEY_CONNECT] = 0
                        logger.info("Disconnect %s and %s"%(link_info.end_infos[0].instance_id,link_info.end_infos[1].instance_id))
                    else:
                        link_info.parameter[PARAMETER_KEY_CONNECT] = 1
                distance = distance_meter(
                    position_map[link_info.end_infos[0].instance_id],
                    position_map[link_info.end_infos[1].instance_id]
                )
                link_info.parameter[PARAMETER_KEY_DELAY] = int(get_propagation_delay_s(distance))*1000000
                cli.put_link_parameter_async(link_info.node_index,link_info.link_id,link_info.parameter)
                
        sleep(step_second)
