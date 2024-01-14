
'''

type ParameterInfo struct {
	Name           string `json:"name"`
	MinVal         int64  `json:"min_val"`
	MaxVal         int64  `json:"max_val"`
	DefinitionFrac int64  `json:"definition_frac"`
}

type IPInfoType struct {
	V4Addr string `json:"v4_addr"`
	V6Addr string `json:"v6_addr"`
}

type EndInfoType struct {
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
}

type LinkBase struct {
	Enabled           bool                     `json:"enabled"`
	CrossMachine      bool                     `json:"cross_machine"`
	SupportParameters map[string]ParameterInfo `json:"support_parameters"`
	Parameter         map[string]int64         `json:"parameter"`
	Config            LinkConfig               `json:"config"`
	NodeIndex         int                      `json:"node_index"`
	EndInfos          [2]EndInfoType           `json:"end_infos"`
}

type LinkConfig struct {
	LinkID         string           `json:"link_id"`
	InitInstanceID [2]string        `json:"init_instance_id"`
	Type           string           `json:"type"`
	InitParameter  map[string]int64 `json:"init_parameter"`
	IPInfos        [2]IPInfoType    `json:"ip_infos"`
}

'''

class Link:
    def __init__(self,link_id:str,instance_id:list[str],parameters:dict[str,int]):
        self.link_id = link_id
        self.instance_id:list[str] = instance_id
        self.parameters:dict[str,int] = parameters

class ISL(Link):
    
    def __init__(self,link_id:str,instance_id:list[str],parameters:dict[str,int],is_inter_orbit:bool):
        Link.__init__(self,link_id,instance_id,parameters)
        self.is_inter_orbit = is_inter_orbit

class GSL(Link):

    def __init__(self,link_id:str,instance_id:list[str],parameters:dict[str,int]):
        Link.__init__(self,link_id,instance_id,parameters)

