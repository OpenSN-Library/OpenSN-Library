import networkx as nx
import pickle
from loguru import logger

class ConstellationGraph(object):

    def __init__(self):
        # 创建有向图
        self.graph = nx.DiGraph()
        self.current_link_id = 0

    def add_link(self, source, destination, weight=1):
        self.graph.add_edge(source, destination, weight=weight, label=self.current_link_id)
        self.current_link_id += 1

    def calculate_shortest_path(self, source, destination):
        return nx.shortest_path(self.graph, source, destination)

    def add_node(self, node_id, ip):
        self.graph.add_node(SatelliteNetworkXNode(node_id, ip))

    def dump_graph(self, file_name="../configuration/constellation_graph.pkl"):
        with open(file_name, 'wb') as f:
            pickle.dump(self.graph, f)

    def loadGraph(self, file_name="../configuration/constellation_graph.pkl"):
        with open(file_name, "rb") as f:
            self.graph = pickle.load(f)


class SatelliteNetworkXNode:
    def __init__(self, node_id, ip):
        self.node_id = node_id
        self.ip = ip

    def __str__(self):
        return "Node: %s, IP: %s" % (self.node_id, self.ip)

    def __eq__(self, other):
        return self.node_id == other.node_id and self.ip == other.ip

    def __hash__(self):
        return hash((self.node_id, self.ip))


def createAndSave():
    cons = ConstellationGraph()

    cons.add_link(SatelliteNetworkXNode("node_0", "172.18.0.1"), SatelliteNetworkXNode("node_1", "172.18.0.3"),
                  weight=1)
    cons.add_link(SatelliteNetworkXNode("node_1", "172.18.0.3"), SatelliteNetworkXNode("node_0", "172.18.0.1"),
                  weight=1)
    cons.add_link(SatelliteNetworkXNode("node_0", "172.18.0.83"), SatelliteNetworkXNode("node_10", "172.18.0.81"),
                  weight=1)
    cons.add_link(SatelliteNetworkXNode("node_10", "172.18.0.81"), SatelliteNetworkXNode("node_0", "172.18.0.83"),
                  weight=1)
    cons.add_link(SatelliteNetworkXNode("node_1", "172.18.0.9"), SatelliteNetworkXNode("node_2", "172.18.0.11"),
                  weight=1)
    cons.add_link(SatelliteNetworkXNode("node_2", "172.18.0.11"), SatelliteNetworkXNode("node_1", "172.18.0.9"),
                  weight=1)
    # 将所有的接口和中心节点连接起来
    cons.add_link(SatelliteNetworkXNode("node_0", "172.18.0.1"), SatelliteNetworkXNode("node_0", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_0", "center"), SatelliteNetworkXNode("node_0", "172.18.0.1"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_0", "172.18.0.83"), SatelliteNetworkXNode("node_0", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_0", "center"), SatelliteNetworkXNode("node_0", "172.18.0.83"), weight=0)

    cons.add_link(SatelliteNetworkXNode("node_1", "172.18.0.3"), SatelliteNetworkXNode("node_1", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_1", "center"), SatelliteNetworkXNode("node_1", "172.18.0.3"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_1", "172.18.0.9"), SatelliteNetworkXNode("node_1", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_1", "center"), SatelliteNetworkXNode("node_1", "172.18.0.9"), weight=0)

    cons.add_link(SatelliteNetworkXNode("node_2", "172.18.0.11"), SatelliteNetworkXNode("node_2", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_2", "center"), SatelliteNetworkXNode("node_2", "172.18.0.11"), weight=0)

    cons.add_link(SatelliteNetworkXNode("node_10", "172.18.0.81"), SatelliteNetworkXNode("node_10", "center"), weight=0)
    cons.add_link(SatelliteNetworkXNode("node_10", "center"), SatelliteNetworkXNode("node_10", "172.18.0.81"), weight=0)

    cons.add_node("node_0", "172.18.0.1")
    cons.add_node("node_0", "172.18.0.83")
    cons.add_node("node_0", "center")

    cons.add_node("node_1", "172.18.0.9")
    cons.add_node("node_1", "172.18.0.3")
    cons.add_node("node_1", "center")

    cons.add_node("node_2", "172.18.0.11")
    cons.add_node("node_2", "center")

    cons.add_node("node_10", "172.18.0.81")
    cons.add_node("node_10", "center")

    path = cons.calculate_shortest_path(SatelliteNetworkXNode("node_0", "center"),
                                        SatelliteNetworkXNode("node_2", "center"))

    # 打印路径
    # for item in path:
    #     print(item)

    # 进行序列化
    cons.dump_graph()


def loadAndTest():
    cons = ConstellationGraph()
    cons.loadGraph()

    path = cons.calculate_shortest_path(SatelliteNetworkXNode("node_0", "center"),
                                        SatelliteNetworkXNode("node_2", "center"))

    # 打印路径
    for item in path:
        print(item)


def writeIntoRRF(host_name, network_list, prefix_list):
    with open(f"../configuration/frr/"
              f"{host_name}.conf", "w") as f:
        full_str = \
            f"""frr version 7.2.1 
frr defaults traditional
hostname {host_name}
log syslog informational
no ipv6 forwarding
service integrated-vtysh-config
!
router ospf
    redistribute connected
"""
        for index in range(len(network_list)):
            full_str += f"\t network {network_list[index]}/{prefix_list[index]} area 0.0.0.0\n"
        full_str += "!\n"
        full_str += "line vty\n"
        full_str += "!\n"
        f.write(full_str)


def GenerateNetworkX(subnet_map_tmp):
    logger.info("start to dijkstra")
    node_list = []
    cons = ConstellationGraph()
    for key_temp in subnet_map_tmp.keys():
        source_node = subnet_map_tmp[key_temp][0]
        target_node = subnet_map_tmp[key_temp][1]
        source_node_id = source_node.node_id
        target_node_id = target_node.node_id
        source_node_ip = source_node.subnet_ip[key_temp]
        target_node_ip = target_node.subnet_ip[key_temp]
        logger.info(f"source_node_id: {source_node_id}: {source_node_ip} "
                    f"<---> target_node_id: {target_node_id}: {target_node_ip}")
        satellite_source_node = SatelliteNetworkXNode(f"{source_node_id}", source_node_ip)
        satellite_dest_node = SatelliteNetworkXNode(f"{target_node_id}", target_node_ip)
        satellite_source_center = SatelliteNetworkXNode(f"{source_node_id}", "center")
        satellite_dest_center = SatelliteNetworkXNode(f"{target_node_id}", "center")
        # 添加两个ip节点之间的连接
        cons.add_link(satellite_source_node, satellite_dest_node, weight=1)
        cons.add_link(satellite_dest_node, satellite_source_node, weight=1)
        # 添加到中心节点的连接
        cons.add_link(satellite_source_node, satellite_source_center, weight=0)
        cons.add_link(satellite_source_center, satellite_source_node, weight=0)
        # 添加到中心节点的连接
        cons.add_link(satellite_dest_node, satellite_dest_center, weight=0)
        cons.add_link(satellite_dest_center, satellite_dest_node, weight=0)
        # 添加到list之中
        if satellite_source_node not in node_list:
            node_list.append(satellite_source_node)
        if satellite_dest_node not in node_list:
            node_list.append(satellite_dest_node)
        if satellite_source_center not in node_list:
            node_list.append(satellite_source_center)
        if satellite_dest_center not in node_list:
            node_list.append(satellite_dest_center)
    for node in node_list:
        cons.add_node(node.node_id, node.ip)
    cons.dump_graph()
    logger.info("dijkstra end")


if __name__ == "__main__":
    # createAndSave()
    loadAndTest()
