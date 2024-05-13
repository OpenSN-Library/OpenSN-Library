import json
class Node:
    
    def __init__(self) -> None:
        self.node_index: int = 0
        self.free_instance: int = 0
        self.is_master_node: bool = False
        self.l_3_addr_v_4: bytes = b''
        self.l_3_addr_v_6: bytes = b''
        self.l_2_addr: bytes = b''
        self.ns_instance_map: dict[str,str] = {}
        self.ns_link_map: dict[str,str] = {}
        self.node_link_device_info: dict[str,int] = {}

def node_from_json(seq: str) -> None:
    node = Node()
    if seq is None :
        return node
    node.__dict__ = json.loads(seq)
    return node