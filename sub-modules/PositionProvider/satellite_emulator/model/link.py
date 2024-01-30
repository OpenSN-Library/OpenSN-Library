import json,uuid
from satellite_emulator.utils.tools import object2dict

class EndInfo:

    def __init__(self) -> None:
        self.instance_id: str = ""
        self.instance_type: str = ""
        self.instance_pid: int = 0
        self.end_node_index: int = 0

class LinkBase:

    def __init__(self) -> None:
        self.link_id: str = ""
        self.end_infos: list[EndInfo] = []
        self.type: str = ""
        self.address_infos: list[dict[str,str]] = []
        self.link_index: int = 0
        self.enabled: bool = False
        self.cross_machine: bool = False
        self.parameter: dict[str,int] = {} 
        self.node_index: int = 0

def link_from_json(seq: str) -> LinkBase:
    dic = json.loads(seq)
    if "end_infos" in dic and dic["end_infos"] is not None:
        new_list = []
        for item in dic["end_infos"]:
            end_info = EndInfo()
            end_info.__dict__ = item
            new_list.append(end_info)
        dic["end_infos"] = new_list
    link_base = LinkBase()
    link_base.__dict__ = dic
    return link_base

def create_new_link(
        node_index1:int,
        instance_id1:str,
        instance_type1:str,
        node_index2:int,
        instance_id2:str,
        instance_type2:str,
        link_type:str,
        address_info1:dict[str,str] = {},
        address_info2:dict[str,str] = {},
        init_parameter:dict[str,int] = {},
    ) -> list[LinkBase]:
    link_id = uuid.uuid4().hex[:8]
    ret_array: list[LinkBase] = []
    node_indexes = [node_index1]
    if node_index2 != node_index1:
        node_indexes.append(node_index2)
    for node_index in node_indexes:
        ret = LinkBase()
        ret.link_id = link_id
        ret.type = link_type
        ret.end_infos = [EndInfo(),EndInfo()]
        ret.end_infos[0].instance_id = instance_id1
        ret.end_infos[0].instance_type = instance_type1
        ret.end_infos[0].end_node_index = node_index1
        ret.end_infos[1].instance_id = instance_id2
        ret.end_infos[1].instance_type = instance_type2
        ret.end_infos[1].end_node_index = node_index2
        ret.address_infos=[address_info1,address_info2]
        ret.cross_machine =  ret.end_infos[0].end_node_index !=  ret.end_infos[1].end_node_index
        ret.parameter = init_parameter
        ret.node_index = node_index
        ret.enabled = False
        ret_array.append(ret)
    return ret_array

def link_to_json(link_base: LinkBase) -> str:
    dic = object2dict(link_base)
    return json.dumps(dic)

def link_parameter_from_json(seq: str) -> dict[str,int]:
    return json.loads(seq)

def link_parameter_to_json(parameter: dict[str,int]) -> str:
    return json.dumps(parameter)