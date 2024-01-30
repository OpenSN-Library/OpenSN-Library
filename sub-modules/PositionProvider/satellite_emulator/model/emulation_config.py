import json
from satellite_emulator.utils.tools import object2dict

class ResourceLimit:
    def __init__(self) -> None:
        self.nano_cpu: int = 0
        self.memory_byte: int = 0

class InstanceTypeConfig:

    def __init__(self) -> None:
        self.image: str = ""
        self.envs: dict[str,str] = {}
        self.resource_limit: ResourceLimit = ResourceLimit()
        
        
class EmulationInfo:
    
    def __init__(self) -> None:
        self.running: bool = False
        self.type_config: dict[str,InstanceTypeConfig] = {}

def emulation_info_from_json(seq: str) -> EmulationInfo:
    dic = json.loads(seq)
    ret = EmulationInfo()
    ret.running = dic["running"]
    for k,v in dic["type_config"]:
        config = InstanceTypeConfig()
        config.image = v["image"]
        config.envs = v["envs"]
        config.resource_limit = ResourceLimit()
        config.resource_limit.__dict__ = v["resource_limit"]
        ret.type_config[k] = config
    return ret



def emulation_info_to_json(config: EmulationInfo) -> str:
    dic = object2dict(config)
    return json.dumps(dic)
