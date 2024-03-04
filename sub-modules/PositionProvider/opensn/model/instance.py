import json
from opensn.utils.tools import object2dict

'''
type ConnectionInfo struct {
	LinkID       string `json:"link_id"`
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	EndNodeIndex int    `json:"end_node_index"`
}
'''

class ConnectionInfo:

    def __init__(self) -> None:
        self.link_id:str = ""
        self.instance_id:str = ""
        self.instance_type:str = ""
        self.end_node_index:str = -1

class ResourceLimit:
    def __init__(self) -> None:
        self.nano_cpu: int = 0
        self.memory_byte: int = 0

class DeviceRequireInfo:

    def __init__(self) -> None:
        self.dev_name: str = ""
        self.need_num: int = 0
        self.is_mutex: bool = False

class Instance:

    def __init__(self) -> None:
        self.instance_id: str = ""
        self.name: str = ""
        self.type: str = ""
        self.image: str = ""
        self.extra: dict[str,str] = {}
        self.device_info: dict[str,DeviceRequireInfo] = {}
        self.resource: ResourceLimit = 0
        self.node_index: int = -1
        self.connections: dict[str,ConnectionInfo] = []
        self.start: bool = False

class InstanceRuntime:

    def __init__(self) -> None:
        instance_id: str = ""
        state: str = ""
        pid: int = 0
        container_id: str = ""

def instance_runtime_from_json(seq: str) -> InstanceRuntime:
    dic = json.loads(seq)
    instance_runtime = InstanceRuntime()
    instance_runtime.__dict__ = dic
    return instance_runtime
    

def instance_runtime_to_json(runtime: InstanceRuntime) -> str:
    return json.dumps(runtime.__dict__)

def instance_from_json(seq: str) -> Instance:
    dic = json.loads(seq)
    if "device_info" in dic.keys() and dic["device_info"] is not None:
        for k,v in dic["device_info"].items():
            info = DeviceRequireInfo()
            info.__dict__ = v
            dic["device_info"][k] = info
    
    if "connections" in dic.keys() and dic["connections"] is not None:
        for k,v in dic["connections"].items():
            connect = ConnectionInfo()
            connect.__dict__ = v
            dic["connections"][k] = connect
    
    resource = ResourceLimit()
    resource.__dict__ = dic["resource"]
    dic["resource"] = resource
    instance = Instance()
    instance.__dict__ = dic
    return instance

def instance_to_json(instance: Instance) -> str:
    dic = object2dict(instance)
    return json.dumps(dic)