
import time,json
from loguru import logger
from node_instance_watcher import NodeInstanceWatcher
from threading import Thread

from dependency_client import \
        redis_client,\
        etcd_client

from const_var import NODE_LIST_KEY

NodeThreadingMap: dict[int,NodeInstanceWatcher] = {}

def parse_node_change(node_list: list[int]):
    node_set:set[int] = set(node_list)
    del_set:set[int] = set()

    for current_node_index in NodeThreadingMap.keys():
        if current_node_index not in node_set:
            del_set.add(current_node_index)

    for del_node_index in del_set:
        NodeThreadingMap[del_node_index].terminate()
        del NodeThreadingMap[del_node_index]

    for node_index in node_set:
        if node_index not in NodeThreadingMap:
            NodeThreadingMap[node_index] = NodeInstanceWatcher(node_index)
            NodeThreadingMap[node_index].start()

        
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