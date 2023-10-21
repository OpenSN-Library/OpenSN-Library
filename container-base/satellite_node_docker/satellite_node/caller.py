import sys
import os
import pexpect
import pexpect as px
from const_var import *
from topology import ConstellationGraph, SatelliteNetworkXNode
from tools import get_node_id_from_env


def LoadNetworkX(config_file_path=NETWORKX_GRAPH_FILE_PATH) -> dict:
    """
    Load the networkx graph from the file
    :parameter config_file_path: the path of the networkx graph file
    :return: the dict of the networkx graph
    """
    result = {}
    cons = ConstellationGraph()
    satellite_number = int(os.getenv("SATELLITE_NUM"))
    # ----------------------- from start node to other nodes ---------------------------
    start_node_id = os.getenv("NODE_ID")
    start_node = SatelliteNetworkXNode(start_node_id, "center")
    cons.loadGraph(config_file_path)
    for id_tmp in range(satellite_number):
        if ("node_" + str(id_tmp)) != start_node_id:
            dest_node = SatelliteNetworkXNode(("node_" + str(id_tmp)), "center")
            path_seq = cons.calculate_shortest_path(start_node, dest_node)
            path_seq_ips = [item for item in path_seq if item.ip != "center"]
            path_seq_ips_output = []
            for index, ip_tmp in enumerate(path_seq_ips):
                if index % 2 == 0:
                    path_seq_ips_output.append(ip_tmp)
            result[(start_node_id, id_tmp, path_seq[-2].ip)] = path_seq_ips_output
    # ----------------------- from start node to other nodes ---------------------------

    # ----------------------- from other nodes to other nodes ---------------------------
    for id_tmp in range(satellite_number):
        new_start_node_id = "node_" + str(id_tmp)
        new_start_node = SatelliteNetworkXNode(new_start_node_id, "center")
        if new_start_node_id == start_node_id:
            continue
        else:
            for id_tmp2 in range(satellite_number):
                new_dest_node_id = "node_" + str(id_tmp2)
                new_dest_node = SatelliteNetworkXNode(new_dest_node_id, "center")
                if new_dest_node_id == new_start_node_id:
                    continue
                else:
                    path_seq = cons.calculate_shortest_path(new_start_node, new_dest_node)
                    path_seq_ips = [item for item in path_seq if item.ip != "center"]
                    path_seq_ips_output = []
                    for index, ip_tmp in enumerate(path_seq_ips):
                        if index % 2 == 0:
                            path_seq_ips_output.append(ip_tmp)
                    result[(new_start_node_id, id_tmp2, path_seq[-2].ip)] = path_seq_ips_output
    # ----------------------- from other nodes to other nodes ---------------------------
    return result


def copy_the_configuration_file() -> None:
    """
    Copy the configuration file to the /etc/frr/frr.conf
    And start the FRR service
    :return:
    """
    # time.sleep(COPY_FRR_CONFIG_TIMER)
    frr_config_file_path = f"/configuration/frr/{os.getenv('NODE_ID')}.conf"
    with open(frr_config_file_path, "r") as reader:
        content = reader.read()
        with open(FRR_CONFIG_FILE_PATH, "w") as writer:
            writer.write(content)
    os.system(FRR_SERVICE_START_CMD)


def call_netlink_exe_simple(container_id, path_of_exe=NETLINK_EXE_FILE_PATH) -> None:
    process = px.spawn(path_of_exe, encoding="utf-8")
    process.logfile_read = sys.stdout
    process.sendline(f"lipsin set container_id {container_id}")
    process.sendline("lipsin pre calculate")
    process.sendline("q")
    process.expect(pexpect.EOF)


def call_netlink_exe_with_send_lines(send_lines_tmp, path_of_exe=NETLINK_EXE_FILE_PATH) -> None:
    for single_send_line in send_lines_tmp:
        process = px.spawn(path_of_exe, encoding="utf-8")
        process.logfile_read = sys.stdout
        process.sendline(single_send_line)
        process.expect(".*from kernel.*")
        process.sendline("q")
        process.expect(pexpect.EOF)


def call_netlink_exe(container_id, path_of_exe=NETLINK_EXE_FILE_PATH) -> None:
    """
    Call the netlink_test_userspace to add the route
    :parameter container_id: the id of the container
    :parameter path_of_exe: the path of the netlink_test_userspace
    :return: None
    """
    process = px.spawn(path_of_exe, encoding="utf-8")
    process.logfile_read = sys.stdout
    result = LoadNetworkX()
    count = 1
    for destination in result.keys():
        send_line = ""
        send_line += "lipsin add route " + str(destination[0])[5:] + "|" + str(destination[1]) + "|" + str(
            destination[2]) + "|" + str(len(result[destination])) + "|"
        for index, node_item in enumerate(result[destination]):
            if index == len(result[destination]) - 1:
                send_line += node_item.ip
            else:
                send_line += node_item.ip + "|"
        process.sendline(send_line)
        process.expect(".*from kernel.*")
        if count % 50 == 0:
            process.sendline("q")
            process.expect(pexpect.EOF)
            process = px.spawn(path_of_exe, encoding="utf-8")
            process.logfile_read = sys.stdout
        count += 1
    # process.sendline(f"lipsin set container_id {container_id}")
    # process.sendline("lipsin pre calculate")
    process.sendline("q")
    process.expect(pexpect.EOF)


def get_send_lines(all_pairs_tmp, ):
    send_lines = []

    # -------------load graph-----------------
    cons = ConstellationGraph()
    config_file_path = f"{NETWORKX_GRAPH_FILE_PATH}"
    cons.loadGraph(config_file_path)
    # -------------load graph-----------------

    for pair in all_pairs_tmp:
        source_node = pair[0]
        destination_node = pair[1]
        start_node = SatelliteNetworkXNode(f"node_{source_node}", "center")
        end_node = SatelliteNetworkXNode(f"node_{destination_node}", "center")
        path_seq = cons.calculate_shortest_path(start_node, end_node)
        path_seq_ips = [item for item in path_seq if item.ip != "center"]
        real_path_seq = []
        for index, ip_tmp in enumerate(path_seq_ips):
            if index % 2 == 0:
                real_path_seq.append(ip_tmp)
        send_line = ""
        send_line += "lipsin add route " + str(source_node) + "|" + str(destination_node) + "|" + \
                     str(path_seq[-2].ip) + "|" + str(len(real_path_seq)) + "|"
        for index, node_item in enumerate(real_path_seq):
            if index == len(real_path_seq) - 1:
                send_line += node_item.ip
            else:
                send_line += node_item.ip + "|"
        send_lines.append(send_line)

    return send_lines


def previous_config():
    """
    Previous config: 1.[call_netlink_exe] 2.[copy_the_configuration_file]
    :return:
    """
    node_id = get_node_id_from_env()
    copy_the_configuration_file()
    # call netlink exe simple to set interface hash and container id
    call_netlink_exe_simple(node_id)


def generate_source_dest_pair(source_node, destination_node_id_list):
    all_pairs = []
    first_pair = (source_node, destination_node_id_list[0])
    all_pairs.append(first_pair)
    new_source = destination_node_id_list[0]
    for index in range(1, len(destination_node_id_list)):
        all_pairs.append((new_source, destination_node_id_list[index]))
    return all_pairs


if __name__ == '__main__':
    # the first item is source node id, the remaining items are destination node id
    source_node_id = 0
    destination_node_list = []
    if len(sys.argv) <= 2:
        print("Please input the source node id and destination node id")
        exit(0)
    else:
        source_node_id = int(sys.argv[1])
        destination_node_list = []
        for index_arg in range(2, len(sys.argv)):
            destination_node_list.append(int(sys.argv[index_arg]))
    # print(source_node_id)
    # print(destination_node_list)
    # print(generate_source_dest_pair(source_node_id, destination_node_list))
    all_send_lines = get_send_lines(generate_source_dest_pair(source_node_id, destination_node_list))
    call_netlink_exe_with_send_lines(all_send_lines)
