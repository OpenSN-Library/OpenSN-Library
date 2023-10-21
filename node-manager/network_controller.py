import time
from multiprocessing import Process
from multiprocessing.connection import Pipe
from loguru import logger
import os
from math import cos, sin, sqrt
from typing import Dict, List, Tuple
from const_var import *
from satellite_node import SatelliteNode
from global_var import networks
from threading import Thread


def generate_submission_list_for_network_object_creation(missions, submission_size: int):
    submission_list = []
    for i in range(0, len(missions), submission_size):
        submission_list.append(missions[i:i + submission_size])
    return submission_list


def network_object_creation_submission(submission, send_pipe):
    for net_id, container_id1, container_id2 in submission:
        network_key = get_network_key(container_id1, container_id2)
        networks[network_key] = Network(net_id,
                                        container_id1,
                                        container_id2,
                                        NETWORK_DELAY,
                                        NETWORK_BANDWIDTH,
                                        NETWORK_LOSS)
        # print(network_key)
    send_pipe.send("finished")


def create_network_object_with_multiple_process(missions, submission_size):
    current_finished_submission_count = 0
    rcv_pipe, send_pipe = Pipe()
    submission_list = generate_submission_list_for_network_object_creation(missions, submission_size)
    logger.info(f"create_network_object_submission_size: {submission_size}")
    for single_submission in submission_list:
        singleThread = Thread(target=network_object_creation_submission, args=(single_submission, send_pipe))
        singleThread.start()
    while True:
        rcv_string = rcv_pipe.recv()
        if rcv_string == "finished":
            current_finished_submission_count += 1
            if current_finished_submission_count == len(submission_list):
                rcv_pipe.close()
                send_pipe.close()
                break


def get_laser_delay_ms(position1: dict, position2: dict) -> int:
    lat1, lon1, hei1 = position1[LATITUDE_KEY], position1[LONGITUDE_KEY], position1[HEIGHT_KEY]
    lat2, lon2, hei2 = position2[LATITUDE_KEY], position2[LONGITUDE_KEY], position2[HEIGHT_KEY]
    x1, y1, z1 = hei1 * cos(lat1) * cos(lon1), hei1 * cos(lat1) * sin(lon1), hei1 * sin(lat1)
    x2, y2, z2 = hei2 * cos(lat2) * cos(lon2), hei2 * cos(lat2) * sin(lon2), hei2 * sin(lat2)
    dist_square = (x2 - x1) ** 2 + (y2 - y1) ** 2 + (z2 - z1) ** 2  # UNIT: m^2
    logger.info(f"distance: {int(sqrt(dist_square))} light speed: {LIGHT_SPEED}")
    # return int(sqrt(dist_square) / LIGHT_SPEED)  # UNIT: ms
    # ZHF MODIFY
    return 0


def get_network_key(container_id1: str, container_id2: str) -> str:
    if container_id1 > container_id2:
        container_id1, container_id2 = container_id2, container_id1
    return container_id1 + container_id2


class ContainerEntrypoint:
    def __init__(self, veth_name: str, container_id: str):
        self.veth_name = veth_name
        self.container_id = container_id


def get_bridge_interface_name(bridge_id: str) -> str:
    full_name = "br-" + bridge_id
    br_interfaces_str = os.popen(
        '''ip l | grep -e "br-" | awk 'BEGIN{FS=": "}{print $2}' ''').read()  # popen与system可以执行指令,popen可以接受返回对象
    interface_list = br_interfaces_str.split('\n')[:-1]
    for interface_name in interface_list:
        if full_name.startswith(interface_name, 0):
            return interface_name
    raise SystemError("Interface Not Found")


def get_vethes_of_bridge(interface_name: str) -> list:
    command = "ip l | " \
              "grep -e \"veth\" | " \
              "grep \"%s\" | " \
              "awk \'BEGIN{FS=\": \"}{print $2}\' | " \
              "awk \'BEGIN{FS=\"@\"}{print $1}\'" % interface_name
    veth_list_str = os.popen(command).read()
    veth_list = veth_list_str.split("\n")[:-1]
    return veth_list


class Network:

    def __init__(self, bridge_id: str,
                 container_id1: str,
                 container_id2: str,
                 time: int,
                 band_width: str,
                 loss_percent: str):
        # 为保证network key的唯一性，设置map中key的字符串拼接顺序为小id在前,大id在后
        self.br_id = bridge_id
        self.br_interface_name = get_bridge_interface_name(bridge_id)
        self.veth_interface_list = get_vethes_of_bridge(self.br_interface_name)
        self.delay = time
        self.bandwidth = band_width
        self.loss = loss_percent
        if len(self.veth_interface_list) != 2:
            logger.warning(self.veth_interface_list)
            raise ValueError("wrong veth number of bridge")
        self.veth_map = {
            container_id1: self.veth_interface_list[0],
            container_id2: self.veth_interface_list[1]
        }
        self.init_info()

    def init_info(self):
        command = "tc qdisc add dev %s %s netem delay %dms loss %s rate %s" % (
            self.veth_interface_list[0], "root", self.delay, self.loss, self.bandwidth)
        exec_res = os.popen(command).read()
        logger.info(f"{self.veth_interface_list[0]} init")
        command = "tc qdisc add dev %s %s netem delay %dms loss %s rate %s" % (
            self.veth_interface_list[1], "root", self.delay, self.loss, self.bandwidth)
        exec_res = os.popen(command).read()
        logger.info(f"{self.veth_interface_list[1]} init")

    def update_info(self):
        command = "tc qdisc replace dev %s %s netem delay %dms loss %s rate %s" % (
            self.veth_interface_list[0], "root", self.delay, self.loss, self.bandwidth)
        exec_res = os.popen(command).read()
        logger.info(f"{self.veth_interface_list[0]} update {self.delay}")
        command = "tc qdisc replace dev %s %s netem delay %dms loss %s rate %s" % (
            self.veth_interface_list[1], "root", self.delay, self.loss, self.bandwidth)
        exec_res = os.popen(command).read()
        logger.info(f"{self.veth_interface_list[1]} update {self.delay}")

    def update_delay_param(self, set_time: int):
        self.delay = set_time

    def update_bandwidth_param(self, band_width: str):
        self.bandwidth = band_width

    def update_loss_param(self, loss_percent: str):
        self.loss = loss_percent


def generate_mission_for_update_network_delay(position_data: Dict[str, Dict[str, float]], topo: Dict[str, List[str]],
                                              satellite_map_tmp: Dict[str, SatelliteNode]):
    update_network_delay_missions = []
    for start_node_id in topo.keys():
        conn_array = topo[start_node_id]
        for target_node_id in conn_array:
            start_container_id = satellite_map_tmp[start_node_id].container_id
            target_container_id = satellite_map_tmp[target_node_id].container_id
            delay = get_laser_delay_ms(position_data[start_node_id], position_data[target_node_id])
            network_key = get_network_key(start_container_id, target_container_id)
            update_network_delay_missions.append(
                (network_key, delay)
            )
    return update_network_delay_missions


def generate_submission_list_for_update_network_delay(missions: List[Tuple[str, int]],
                                                      submission_size: int):
    submission_list = []
    for i in range(0, len(missions), submission_size):
        submission_list.append(missions[i:i + submission_size])
    return submission_list


def update_network_delay_with_single_process(submission: List[Tuple[str, int]], networks_tmp, send_pipe):
    for network_key, delay in submission:
        network = networks_tmp[network_key]
        network.update_delay_param(delay)
        network.update_info()
    send_pipe.send("finished")


def update_network_delay_with_multi_process(stop_process_state,
                                            networks_tmp,
                                            position_data: Dict[str, Dict[str, float]],
                                            topo: Dict[str, List[str]],
                                            satellite_map_tmp: Dict[str, SatelliteNode],
                                            submission_size: int,
                                            update_interval: int):
    # update count
    update_count = 0
    # generate missions
    missions = generate_mission_for_update_network_delay(position_data, topo, satellite_map_tmp)
    # generate submission list
    submission_list = generate_submission_list_for_update_network_delay(missions, submission_size)
    # submit
    while True:
        # store the start time
        start_time = time.time()
        if stop_process_state.value:
            break
        # current count
        current_count = 0
        # generate pipe
        rcv_pipe, send_pipe = Pipe()
        for submission in submission_list:
            # process = Process(target=update_network_delay_with_single_process,
            #                   args=(submission, networks_tmp, send_pipe))
            # process.start()
            singleThread = Thread(target=update_network_delay_with_single_process,
                                  args=(submission, networks_tmp, send_pipe))
            singleThread.start()
        # receive the result
        while True:
            rcv_string = rcv_pipe.recv()
            if rcv_string == "finished":
                current_count += 1
                # traverse all the process and kill them
                if current_count == len(submission_list):
                    send_pipe.close()
                    rcv_pipe.close()
                    break
        end_time = time.time()
        logger.info(f"update satellite network delay in {end_time - start_time}s")
        update_count += 1
        if update_count == 1:
            break
        time.sleep(update_interval)
    logger.success("update satellite network delay process finished")


if __name__ == '__main__':
    print(generate_submission_list_for_network_object_creation([1, 2, 3, 4, 5], 1))
