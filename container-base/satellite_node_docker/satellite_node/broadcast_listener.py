import multiprocessing
import random
import socket
import logging
import json
import os
import math
import time
from multiprocessing import Process, Pipe
from laser_simulator import judge_connect
from data_updater import DataUpdater
from const_var import *
from caller import previous_config
from tools import get_node_id_from_env


def read_interface_from_config():
    interface_map = {}
    current_node_id = get_node_id_from_env()
    with open(f"/configuration/interface_table/node{current_node_id}_interface_table.conf", "r") as f:
        for line in f.readlines():
            interface = line.strip().split(":")
            interface_map[interface[0]] = interface[1]
    return interface_map


def broadcast_with_timer(updater, status_string, time_to_sleep):
    time.sleep(time_to_sleep)
    updater.broadcast_status(status_string)


def listen_worker(port, buffer_size, node_id, threshold, updater) -> None:
    """
    Listen the broadcast message
    :param port: the port to listen
    :param buffer_size: the buffer size
    :param node_id: the current node id
    :param updater: the link status sender
    :param threshold: the latitude threshold to judge the connection
    :return: None
    """
    udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    udp_socket.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
    udp_socket.bind(('0.0.0.0', port))
    logging.info("Listening on port %d" % port)
    while True:
        json_data, client_address = udp_socket.recvfrom(buffer_size)
        # decode the message to object
        data = json.loads(str(json_data, encoding="utf-8"))
        # when having the config key in the data
        if "config" in data:
            if data["config"] == "set the source routing table":
                interface_map = read_interface_from_config()
                # print(interface_map, flush=True)
                previous_config()
            continue
        # get the current satellite node data
        # print(data)
        self_data = data['position_datas'][node_id]
        latitude = self_data[LATITUDE_KEY] * 180 / math.pi
        longitude = self_data[LONGITUDE_KEY] * 180 / math.pi
        height = self_data[HEIGHT_KEY]
        logging.info("self.node_id is %s, self.Latitude is %f, Longitude is %f, Sea level height is %f" % (
            node_id, latitude, longitude, height))
        # judge the connection state
        next_conn_state = judge_connect(latitude, threshold)

        # change the interface state
        if next_conn_state:
            # execute command to change the interface state
            for interface in interface_map.keys():
                if interface_map[interface] == "inter-orbit-link":
                    command = f"ifconfig {interface} up"
                    logging.info(f"Exec '{command}'")
                    os.system(command)
            status_str = json.dumps({
                "node_id": node_id,
                "state": True
            })
            # broadcast the status to the backend
            random_time_to_sleep_before_broadcast = random.uniform(0, 1)
            process = Process(target=broadcast_with_timer,
                              args=(updater, status_str, random_time_to_sleep_before_broadcast))
            process.start()
        else:
            for interface in interface_map.keys():
                if interface_map[interface] == "inter-orbit-link":
                    command = "ifconfig %s down" % interface
                    logging.info(f"Exec '{command}'")
                    os.system(command)
            status_str = json.dumps({
                "node_id": node_id,
                "state": False
            })
            # broadcast the status to the backend
            random_time_to_sleep_before_broadcast = random.uniform(0, 1)
            process = Process(target=broadcast_with_timer,
                              args=(updater, status_str, random_time_to_sleep_before_broadcast))
            process.start()


def satellite_simulator(broad_port: int, node_id: str, threshold: float):
    """
    The main satellite simulator process
    :param broad_port:  the broadcast port
    :param node_id:  the current node id
    :param threshold:  the latitude threshold to judge the connection
    :return:
    """
    broad_receive_port = broad_port
    broad_send_port = broad_port + 1
    monitor_ip = os.getenv("MONITOR_IP")
    updater = DataUpdater(monitor_ip, broad_send_port)
    pipe_recv, pipe_send = Pipe()
    # create listen_worker process and start
    Process(target=listen_worker, args=(broad_receive_port, BUFFER_SIZE, node_id, threshold, updater)).start()


if __name__ == "__main__":
    print(list(range(2, 12, 2)))
