
'''
type LinkInfo struct {
	V4Addr string `json:"v4_addr"`
	V6Addr string `json:"v6_addr"`
}

type EndInfo struct {
	InstanceID string `json:"instance_id"`
	Type       string `json:"type"`
}

type TopoInfo struct {
	InstanceID string              `json:"instance_id"`
	LinkInfos  map[string]*LinkInfo `json:"link_infos"`
	EndInfos   map[string]*EndInfo  `json:"end_infos"`
}
'''

class EndInfo:

    def __init__(self,instance_id:str, type:str) -> None:
        self.instance_id = instance_id
        self.type = type

class LinkInfo:

    def __init__(self,v4_addr:str,v6_addr:str) -> None:
        self.v4_addr = v4_addr
        self.v6_addr = v6_addr

class InstancConfig:

    def __init__(self,instance_id:str) -> None:
        self.instance_id = instance_id
        self.link_infos: dict[str,LinkInfo] = {}
        self.end_infos:dict[str,EndInfo] = {}
