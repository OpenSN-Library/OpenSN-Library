from satellite_emulator.model.node import Node,node_from_json
from satellite_emulator.const.etcd_key import NODE_LIST_KEY
from etcd3 import Etcd3Client

def get_node_map(etcd_client:Etcd3Client,) -> dict[int,Node]:
    ret: dict[int,Node] = {}
    resps = etcd_client.get_prefix(NODE_LIST_KEY)
    for val,meta in resps:
        ret[int(meta.key.decode().split('/')[-1])] = node_from_json(val)
    return ret

def get_node(etcd_client:Etcd3Client,node_index) -> Node:
    node_key = "%s/%d"%(NODE_LIST_KEY,node_index)
    val,meta = etcd_client.get(node_key)
    return node_from_json(val)