import networkx as nx
import pickle


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

    def dump_graph(self, file_name="/configuration/constellation_graph.pkl"):
        with open(file_name, 'wb') as f:
            pickle.dump(self.graph, f)

    def loadGraph(self, file_name="/configuration/constellation_graph.pkl"):
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


def loadAndTest():
    cons = ConstellationGraph()
    cons.loadGraph()

    path = cons.calculate_shortest_path(SatelliteNetworkXNode("node_0", "center"),
                                        SatelliteNetworkXNode("node_5", "center"))

    # 打印路径
    for item in path:
        print(item)


if __name__ == "__main__":
    loadAndTest()
