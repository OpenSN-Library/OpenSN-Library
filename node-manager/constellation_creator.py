import time
from collections import OrderedDict
from threading import Thread

import docker
from tools import ip_to_subnet
from satellite_node import SatelliteNode
from const_var import *
from topology import writeIntoRRF, GenerateNetworkX
from network_controller import Network, get_network_key
from loguru import logger
from multiprocessing import Process, Pipe
from queue import PriorityQueue
from subnet_allocator import ip2str
from global_var import networks, satellites, connect_order_map, satellite_map
from network_controller import create_network_object_with_multiple_process
from global_var import interface_map_lock, interface_map


def create_satellite_submission(mission_index,
                                mission,
                                docker_client,
                                udp_listening_port,
                                satellite_num,
                                successful_init,
                                send_pipe):
    """
    create satellite submission
    :param mission_index: submission index
    :param mission: assigned submission
    :param docker_client: docker client
    :param udp_listening_port: udp listening port
    :param satellite_num: satellite number
    :param successful_init: successful init
    :param send_pipe: send pipe (we need to put the container id into the pipe)
    :return:
    """
    for item in mission:
        index = item - 1
        node_id = 'node_' + str(index)
        # after docker_client create satellite the container_id will be returned
        container_id = docker_client.create_satellite(
            node_id,
            udp_listening_port,
            satellite_num,
            successful_init,
        )
        logger.success(f"create satellite {node_id} successfully")
        send_pipe.send(f"{mission_index}|{index}|{container_id}")
    send_pipe.send("finished")


def generate_submission_list(satellite_number: int, submission_size_tmp: int):
    """
    generate submission for create satellite container
    :param satellite_number: total satellite container need to create
    :param submission_size_tmp: the number of the satellite container that each submission should create
    :return: mission list and container id list
    """
    mission_list = []
    mission_list_tmp = []
    for i in range(1, satellite_number + 1):
        # submission_size_tmp satellites in one mission
        if i % submission_size_tmp == 0:
            mission_list_tmp.append(i)
            mission_list.append(mission_list_tmp)
            mission_list_tmp = []
        else:
            mission_list_tmp.append(i)
            if i == satellite_number:
                mission_list.append(mission_list_tmp)
    return mission_list


def print_and_store_interface_map(interface_map_tmp):
    """
    print and store interface map
    :param interface_map_tmp: interface map [key:node_id][value:interface list]
    :return:
    """
    for node_id in interface_map_tmp.keys():
        logger.success(f"node id: {node_id} interfaces: {interface_map_tmp[node_id]['interface']}")
    for node_id in interface_map_tmp.keys():
        storage = f"../configuration/interface_table" \
                  f"/node{node_id}_interface_table.conf"
        with open(storage, "w") as f:
            for interface in interface_map_tmp[node_id]["interface"]:
                f.write(f"{interface[0]}:{interface[1]}\n")


def modify_interface_map(conn_index, network_node_id, interface_map_tmp):
    """
    modify interface map
    :param conn_index:  start node id
    :param network_node_id:  end node id
    :param interface_map_tmp:  interface map that need to be modified
    :return:
    """
    # 打印两个点所处的轨道
    start_node_id = conn_index
    end_node_id = str(network_node_id)
    start_node_orbit = int(conn_index) // SAT_PER_ORBIT
    end_node_orbit = int(network_node_id) // SAT_PER_ORBIT
    # 如果两个点在同一个轨道上
    if start_node_orbit == end_node_orbit:
        # print(f"start node id:{conn_index} end node id:{network_node_id} intra-orbit-link", flush=True)
        # operation for start node
        if start_node_id not in interface_map_tmp.keys():
            interface_map_tmp[start_node_id] = {"current_index": 2, "interface": [("eth1", "intra-orbit-link")]}
        else:
            interface_map_tmp[start_node_id]["interface"]. \
                append((f"eth{interface_map_tmp[start_node_id]['current_index']}", "intra-orbit-link"))
            interface_map_tmp[start_node_id]["current_index"] += 1
        # operation for end node
        if end_node_id not in interface_map_tmp.keys():
            interface_map_tmp[end_node_id] = {"current_index": 2, "interface": [("eth1", "intra-orbit-link")]}
        else:
            interface_map_tmp[end_node_id]["interface"].append(
                (f"eth{interface_map_tmp[end_node_id]['current_index']}", "intra-orbit-link"))
            interface_map_tmp[end_node_id]["current_index"] += 1
    # 如果两个点不在同一个轨道上
    else:
        # print(f"start node id:{start_node_id} end node id:{network_node_id} inter-orbit-link", flush=True)
        # operation for start node
        if start_node_id not in interface_map_tmp.keys():
            interface_map_tmp[start_node_id] = {"current_index": 2, "interface": [("eth1", "inter-orbit-link")]}
        else:
            interface_map_tmp[start_node_id]["interface"]. \
                append((f"eth{interface_map_tmp[start_node_id]['current_index']}", "inter-orbit-link"))
            interface_map_tmp[start_node_id]["current_index"] += 1
        # operation for end node
        if end_node_id not in interface_map_tmp.keys():
            interface_map_tmp[end_node_id] = {"current_index": 2, "interface": [("eth1", "inter-orbit-link")]}
        else:
            interface_map_tmp[end_node_id]["interface"].append(
                (f"eth{interface_map_tmp[end_node_id]['current_index']}", "inter-orbit-link"))
            interface_map_tmp[end_node_id]["current_index"] += 1


def multiprocess_generate_containers(submission_size_tmp,
                                     satellite_num,
                                     docker_client,
                                     udp_listening_port,
                                     successful_init,
                                     satellite_infos,
                                     position_datas,
                                     satellites_tmp):
    submission_size = submission_size_tmp  # the size of the submission
    finished_submission_count = 0  # the count of the finished submission
    rcv_pipe, send_pipe = Pipe()  # the pipe for receiving the container_id
    satellite_id_to_satellite = PriorityQueue()  # the priority queue for the (satellite_id, satellite) pair
    mission_list = generate_submission_list(satellite_num, submission_size)
    process_list = []
    logger.info(f"container creation submission list size: {len(mission_list)}")
    for mission_index, mission in enumerate(mission_list):
        process = Process(target=create_satellite_submission,
                          args=(mission_index,
                                mission,
                                docker_client,
                                udp_listening_port,
                                satellite_num,
                                successful_init,
                                send_pipe))
        process_list.append(process)
        process.start()

    # collect the generated container_id
    while True:
        rcv_string = rcv_pipe.recv()
        if rcv_string == "finished":
            finished_submission_count += 1
            if finished_submission_count == len(mission_list):
                rcv_pipe.close()
                send_pipe.close()
                break
        else:
            mission_index, node_id, container_id = rcv_string.split("|")
            node_id_str = "node_" + str(node_id)
            tmp_satellite_node = SatelliteNode((
                satellite_infos[int(node_id)][0],
                satellite_infos[int(node_id)][1],
                satellite_infos[int(node_id)][2],
            ), node_id_str, container_id)
            satellite_map[tmp_satellite_node.node_id] = tmp_satellite_node
            satellite_id_to_satellite.put((int(node_id), tmp_satellite_node))
            position_datas[node_id_str] = {
                LATITUDE_KEY: 0.0,
                LONGITUDE_KEY: 0.0,
                HEIGHT_KEY: 0.0
            }

    # terminate the process
    for process in process_list:
        process.kill()

    # construct satellites
    while not satellite_id_to_satellite.empty():
        _, satellite_node = satellite_id_to_satellite.get()
        satellites_tmp.append(satellite_node)


def modify_interface_map_process(link_connections, interface_map_tmp):
    for conn_index in link_connections.keys():
        # network_node_id is the end node
        for network_node_id in link_connections[conn_index]:
            modify_interface_map(conn_index, network_node_id, interface_map_tmp)
    print_and_store_interface_map(interface_map_tmp)


def generate_mission_for_network(link_connections, satellites_tmp, docker_client):
    # final result
    final_mission_list = []
    # network_index
    network_index = 0
    # conn_index is the start node str
    for conn_index in link_connections.keys():
        # network_node_id is the end node
        for network_node_id in link_connections[conn_index]:
            # using the allocator the allocate the network
            subnet_ip = docker_client.allocator.alloc_local_subnet()
            subnet_ip_str = ip2str(subnet_ip)
            gateway_str = ip2str(subnet_ip + 1)
            ipam_pool = docker.types.IPAMPool(subnet='%s/29' % subnet_ip_str, gateway=gateway_str)
            ipam_config = docker.types.IPAMConfig(pool_configs=[ipam_pool])
            # start node id
            node_id1 = satellites_tmp[int(conn_index)].node_id
            # end node id
            node_id2 = satellites_tmp[network_node_id].node_id
            # update connect order map
            if node_id1 in connect_order_map.keys():
                connect_order_map[node_id1].append(node_id2)
            else:
                connect_order_map[node_id1] = [node_id2]
            # start container id
            container_id1 = satellite_map[node_id1].container_id
            # end container id
            container_id2 = satellite_map[node_id2].container_id
            # tmp mission generated
            tmp_mission = (node_id1, node_id2, container_id1, container_id2, network_index, ipam_config)
            # add the tmp mission into submission list
            final_mission_list.append(tmp_mission)
            # update network index
            network_index += 1
    return final_mission_list


def generate_submission_list_for_network(missions, submission_size_tmp):
    """
    generate submission list for network
    :param missions:  mission list to create docker network
    :param submission_size_tmp:  submission size
    :return:
    """
    submission_list = []
    submission_list_tmp = []
    mission_count = len(missions)
    for i in range(1, mission_count + 1):
        # submission_size_tmp satellites in one mission
        if i % submission_size_tmp == 0:
            submission_list_tmp.append(missions[i - 1])
            submission_list.append(submission_list_tmp)
            submission_list_tmp = []
        else:
            submission_list_tmp.append(missions[i - 1])
            if i == mission_count:
                submission_list.append(submission_list_tmp)
    return submission_list


def create_network_submission(submission, docker_client, send_pipe):
    for single_mission in submission:
        # split and get the mission info
        node_id1, node_id2, container_id1, container_id2, network_index, ipam_config = single_mission
        # create network
        net_id = docker_client.create_network_with_ipam_config(network_index, ipam_config)
        # connect the node to the network
        docker_client.connect_node(container_id1, net_id, node_id1)
        docker_client.connect_node(container_id2, net_id, node_id2)
        # print link connection information
        logger.info("connect satellite %s and %s" % (node_id1, node_id2))
        # send the net_id info
        send_pipe.send(f"{net_id}|{container_id1}|{container_id2}")
        # modify interface map
        interface_map_lock.acquire()
        modify_interface_map(node_id1[5:], node_id2[5:], interface_map)
        interface_map_lock.release()
    # after all the network is created
    send_pipe.send("finished")


def multiple_process_generate_networks(link_connections, satellites_tmp, docker_client, submission_size):
    # create pipe
    rcv_pipe, send_pipe = Pipe()
    # generate networks with multiple process
    all_network_missions = generate_mission_for_network(link_connections, satellites_tmp, docker_client)
    for mission in all_network_missions:
        logger.info(mission)
    # generate submission list
    submission_list = generate_submission_list_for_network(all_network_missions, submission_size)
    # print the list length
    logger.info("generate network submission list length: %s" % len(submission_list))
    # total submission count
    submission_count = len(submission_list)
    # finished submission count
    finished_submission_count = 0
    # network object mission
    network_object_mission = []
    # process list
    process_list = []
    # traverse the submission_list
    for submission in submission_list:
        # print(submission)
        # process = Process(target=create_network_submission, args=(submission, docker_client, send_pipe))
        # process_list.append(process)
        # process.start()
        singleThread = Thread(target=create_network_submission, args=(submission, docker_client, send_pipe))
        singleThread.start()
    while True:
        pipe_rcv_string = rcv_pipe.recv()
        if pipe_rcv_string == "finished":
            finished_submission_count += 1
            if finished_submission_count == submission_count:
                send_pipe.close()
                rcv_pipe.close()
                break
        else:
            net_id, container_id1, container_id2 = pipe_rcv_string.split("|")
            network_object_mission.append((net_id, container_id1, container_id2))

    print_and_store_interface_map(interface_map)

    logger.info(f"generate network object mission size: {len(network_object_mission) / SUBMISSION_SIZE_FOR_NETWORK_OBJECT_CREATION}")
    start_time = time.time()
    create_network_object_with_multiple_process(network_object_mission, SUBMISSION_SIZE_FOR_NETWORK_OBJECT_CREATION)
    end_time = time.time()
    logger.success(f"network initialize time: {end_time - start_time} s")
    # for net_id, container_id1, container_id2 in network_object_mission:
    #     network_key = get_network_key(container_id1, container_id2)
    #     # this can be process in multiple process
    #     networks[network_key] = Network(net_id,
    #                                     container_id1,
    #                                     container_id2,
    #                                     NETWORK_DELAY,
    #                                     NETWORK_BANDWIDTH,
    #                                     NETWORK_LOSS)


def constellation_creator(docker_client,
                          satellite_infos,
                          link_connections,
                          host_ip,
                          udp_listening_port,
                          successful_init):
    """
    create constellation function
    :param docker_client:  docker client
    :param satellite_infos:  satellite info list
    :param link_connections:  link connection list
    :param host_ip:  host ip
    :param udp_listening_port:  udp listening port
    :param successful_init:  successful init monitor
    :return:
    """
    # interface_map = {}  # 接口表 interface map
    # satellites = []  # 卫星列表 satellite list
    position_datas = {}  # 位置信息表 position data map from node_id to position data
    subnet_map = OrderedDict()  # 子网表 subnet map ordered dict the order is the order of appending
    network_index = 0  # 网络节点索引 network node index
    satellite_num = len(satellite_infos)  # 卫星数量 satellite number
    host_subnet = ip_to_subnet(host_ip, HOST_PREFIX_LEN)  # 主机子网 host subnet
    submission_size_for_container = SUBMISSION_SIZE_FOR_CONTAINER_CREATION  # the size of the submission for container
    submission_size_for_network = SUBMISSION_SIZE_FOR_NETWORK_CREATION  # the size of the submission for network

    # generate containers with multiple process
    # ---------------------------------------------------------------------
    multiprocess_generate_containers(submission_size_for_container,
                                     satellite_num,
                                     docker_client,
                                     udp_listening_port,
                                     successful_init,
                                     satellite_infos,
                                     position_datas,
                                     satellites)
    # ---------------------------------------------------------------------

    # modify interface table
    # ---------------------------------------------------------------------
    # modify_interface_map_process(link_connections, interface_map)
    # ---------------------------------------------------------------------

    # generate containers with multiple process
    # ---------------------------------------------------------------------
    multiple_process_generate_networks(link_connections, satellites, docker_client, submission_size_for_network)
    # ---------------------------------------------------------------------

    # used the actual link connection to modify interface map
    # ---------------------------------------------------------------------

    # ---------------------------------------------------------------------

    # network generate process
    # ---------------------------------------------------------------------
    """
    # conn_index is the start node
    for conn_index in link_connections.keys():
        # network_node_id is the end node
        for network_node_id in link_connections[conn_index]:
            net_id = docker_client.create_network(network_index)
            network_index += 1
            node_id1 = satellites[int(conn_index)].node_id
            node_id2 = satellites[network_node_id].node_id
            container_id1 = satellite_map[node_id1].container_id
            container_id2 = satellite_map[node_id2].container_id
            docker_client.connect_node(container_id1, net_id, node_id1)
            docker_client.connect_node(container_id2, net_id, node_id2)
            if node_id1 in connect_order_map.keys():
                connect_order_map[node_id1].append(node_id2)
            else:
                connect_order_map[node_id1] = [node_id2]
            # print link connection information
            print("connect satellite %s and %s" % (node_id1, node_id2))
            # create the network object, this object is used to set bandwidth and delay
            network_key = get_network_key(container_id1, container_id2)
            networks[network_key] = Network(net_id, container_id1, container_id2)
            # set the delay bandwidth and loss rate
            networks[network_key].update_bandwidth(NETWORK_BANDWIDTH)
            networks[network_key].update_delay(NETWORK_DELAY)
            networks[network_key].update_loss(NETWORK_LOSS)
    """

    # traverse all the satellite
    for container_id_key in satellite_map.keys():
        # get satellite
        container_id = satellite_map[container_id_key].container_id
        # get interface and prefix length of container
        interfaces, prefix_len = docker_client.get_container_interfaces(container_id)
        toDelete = -1.
        # traverse all the interfaces and delete the host interface
        for i in range(len(interfaces)):
            if ip_to_subnet(interfaces[i], HOST_PREFIX_LEN) == host_subnet:
                toDelete = i
                break
        if toDelete >= 0:
            satellite_map[container_id_key].host_ip = interfaces[toDelete]
            del interfaces[toDelete]
            del prefix_len[toDelete]
        # get the subnet of each interface
        sub_nets = [ip_to_subnet(interfaces[i], prefix_len[i]) for i in range(len(prefix_len))]
        # write the interface into the rrf file
        writeIntoRRF(container_id_key, sub_nets, prefix_len)
        # traverse all the subnets
        for sub_index in range(len(sub_nets)):
            # sub is the subnet
            sub = sub_nets[sub_index]
            # subnet_ip is a map {subnet_str->interface}
            satellite_map[container_id_key].subnet_ip[sub] = interfaces[sub_index]
            # if the subnet is in the subnet map
            if sub in subnet_map.keys():
                # append the satellite into the subnet map
                subnet_map[sub].append(satellite_map[container_id_key])
            else:
                # create a new subnet
                subnet_map[sub] = [satellite_map[container_id_key]]

    # 进行每一条星间链路的打印
    GenerateNetworkX(subnet_map)
    # 所有的星间链路都存储在subnet_map中，这里进行拓扑的构建
    """
    original code
    for k in subnet_map.keys():
        subnet_map[k][0].topo.append({
            'source_ip': subnet_map[k][0].subnet_ip[k],
            'target_ip': subnet_map[k][1].subnet_ip[k],
            'target_node_id': subnet_map[k][1].node_id
        })
    """
    # subnet_map中的每一个key都是一个子网，对于每一个子网，都有两个卫星，这两个卫星的连接关系是
    # subnet_map[k][0]和subnet_map[k][1]，这里需要对这两个卫星的连接关系进行判断，如果
    # subnet_map[k][1]在subnet_map[k][0]的连接顺序中，那么就需要进行交换
    for k in subnet_map.keys():
        if subnet_map[k][1].node_id in connect_order_map.keys() \
                and subnet_map[k][0].node_id in connect_order_map[subnet_map[k][1].node_id]:
            temp = subnet_map[k][0]
            subnet_map[k][0] = subnet_map[k][1]
            subnet_map[k][1] = temp
        # if same orbit
        if subnet_map[k][0].orbit == subnet_map[k][1].orbit:
            same_orbit = True
        else:
            same_orbit = False
        if same_orbit:
            subnet_map[k][0].topo.insert(0, {
                'source_ip': subnet_map[k][0].subnet_ip[k],
                'target_ip': subnet_map[k][1].subnet_ip[k],
                'target_node_id': subnet_map[k][1].node_id
            })
        else:
            subnet_map[k][0].topo.append({
                'source_ip': subnet_map[k][0].subnet_ip[k],
                'target_ip': subnet_map[k][1].subnet_ip[k],
                'target_node_id': subnet_map[k][1].node_id
            })

    # zhf add code
    monitor_payloads_queue = PriorityQueue()
    for k in satellite_map.keys():
        monitor_payloads_queue.put((int(satellite_map[k].node_id[5:]),
                                    satellite_map[k].node_id,
                                    satellite_map[k].host_ip,
                                    satellite_map[k].topo))
    monitor_payloads = []
    while not monitor_payloads_queue.empty():
        current = monitor_payloads_queue.get()
        monitor_payloads.append({
            "node_id": current[1],
            "host_ip": current[2],
            "connections": current[3]
        })

    return [position_datas, monitor_payloads]


if __name__ == "__main__":
    list_tmp = generate_submission_list_for_network([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 3)
    print(list_tmp)
