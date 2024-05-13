import json


'''
type TypeConfigReq struct {
	Image         string            `json:"image"`
	Envs          map[string]string `json:"container_envs"`
	ResourceLimit ResourceLimitStr  `json:"resource_limit"`
}
'''

class EmulationTypeConfig :
    def __init__(self,image:str,envs:dict[str,str],nano_cpu:str,mem_byte:str) -> None:
        self.image: str = image
        self.container_envs: dict[str,str] = envs
        self.resource_limit: dict[str,int] = {
            "nano_cpu": nano_cpu,
            "memory_byte": mem_byte
        }

class TopologyLink:

    def __init__(self,type: str,end_indexes: list[int],init_parameter: dict[str,int],address_info: list[dict[str,str]],extra: dict[str,str]) -> None:
        self.end_indexes: list[int] = end_indexes
        self.type: str = type
        self.init_parameter: dict[str,int] = init_parameter
        self.address_infos: list[dict[str,str]] = address_info
        self.extra: dict[str,str] = extra

class TopologyInstance:
    
    def __init__(self,type: str,extra: dict[str,str]) -> None:
        self.type: str = type
        self.extra: dict[str,str] = extra
        self.device_info = {}

class TopologyConfig:

    def __init__(self) -> None:
        self.instances : list[TopologyInstance] = []
        self.links : list[TopologyLink] = []

    def toJson(self):
        cp = TopologyConfig()
        for instance in self.instances:
            cp.instances.append(instance.__dict__)
        for link in self.links:
            cp.links.append(link.__dict__)
        return json.dumps(cp.__dict__)