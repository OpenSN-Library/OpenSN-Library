
import time,json
from loguru import logger
from node_instance_watcher import NodeInstanceWatcher
from node_link_watcher import NodeLinkWatcher
from threading import Thread

from dependency_client import \
        redis_client,\
        etcd_client

from const_var import NODE_LIST_KEY

NodeInstanceThreadingMap: dict[int,NodeInstanceWatcher] = {}
NodeLinkThreadingMap: dict[int,NodeLinkWatcher] = {}

def parse_node_change(node_list: list[int]):
    node_set:set[int] = set(node_list)
    del_set:set[int] = set()

    for current_node_index in NodeInstanceThreadingMap.keys():
        if current_node_index not in node_set:
            del_set.add(current_node_index)

    for del_node_index in del_set:
        NodeInstanceThreadingMap[del_node_index].terminate()
        del NodeInstanceThreadingMap[del_node_index]

    for node_index in node_set:
        if node_index not in NodeInstanceThreadingMap:
            NodeInstanceThreadingMap[node_index] = NodeInstanceWatcher(node_index)
            NodeInstanceThreadingMap[node_index].start()
            NodeLinkThreadingMap[node_index] = NodeLinkWatcher(node_index)
            NodeLinkThreadingMap[node_index].start()

            

        
class NodeListWatcher(Thread):

    def __init__(self):
        Thread.__init__(self)
        self.cancel = None
        self.stop_sig = True

    def terminate(self):
        if self.cancel is not None:
            self.cancel()
            self.cancel = None
        self.stop_sig = True

    def run(self):
        if not self.stop_sig:
            return
        self.stop_sig = False
        val,useless = etcd_client.get(NODE_LIST_KEY)
        if val is not None:
            node_list = json.loads(val)
            parse_node_change(node_list)
        while not self.stop_sig:
            try:
                events,cancel = etcd_client.watch(NODE_LIST_KEY)
                for event in events:
                    node_list = json.loads(event.value)
                    parse_node_change(node_list)
            except Exception as e:
                logger.error("Watch Node List Error %s"%str(e))
                cancel()
                time.sleep(10)