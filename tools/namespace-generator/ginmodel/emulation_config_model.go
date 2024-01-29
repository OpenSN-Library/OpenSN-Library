package ginmodel

type ConfigEmulationReq map[string]InstanceTypeConfig

type EmulationDetail struct {
	InstanceTypeConfig map[string]InstanceTypeConfig `json:"instance_type_config"`
	Running            bool                          `json:"running"`
}

type EmulationConfig struct {
	InstanceTypeConfig map[string]InstanceTypeConfig `json:"instance_type_config"`
	Running            bool                          `json:"running"`
}

type TopologyLink struct {
	EndIndexes    [2]int               `json:"end_indexes"`
	Type          string               `json:"type"`
	InitParameter map[string]int64     `json:"init_parameter"`
	AddressInfos  [2]map[string]string `json:"address_infos"`
}

type TopologyInstance struct {
	Type       string                       `json:"type"`
	Extra      map[string]string            `json:"extra"`
	DeviceInfo map[string]DeviceRequireInfo `json:"device_need"`
}

type AddTopologyReq struct {
	Instances []TopologyInstance `json:"instances"`
	Links     []TopologyLink     `json:"links"`
}

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}
