package ginmodel

type ConfigEmulationReq map[string]TypeConfigReq

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}

type ResourceLimit struct {
	NanoCPU    int64 `json:"nano_cpu"`
	MemoryByte int64 `json:"memory_byte"`
}

type InstanceTypeConfig struct {
	Image         string            `json:"image"`
	Envs          map[string]string `json:"container_envs"`
	ResourceLimit ResourceLimit     `json:"resource_limit"`
}

type EmulationInfo struct {
	Running           bool                          `json:"running"`
	TypeConfig        map[string]InstanceTypeConfig `json:"type_config"`
}
type TypeConfigReq struct {
	Image         string            `json:"image"`
	Envs          map[string]string `json:"container_envs"`
	ResourceLimit ResourceLimitStr  `json:"resource_limit"`
}

type ResourceLimitStr struct {
	NanoCPU    string `json:"nano_cpu"`
	MemoryByte string `json:"memory_byte"`
}

type EmulationDetail struct {
	InstanceTypeConfig map[string]InstanceTypeConfig `json:"instance_type_config"`
	Running            bool                                `json:"running"`
}

type EmulationConfig struct {
	InstanceTypeConfig map[string]InstanceTypeConfig `json:"instance_type_config"`
	Running            bool                                `json:"running"`
}

type TopologyLink struct {
	EndIndexes    [2]int               `json:"end_indexes"`
	Type          string               `json:"type"`
	InitParameter map[string]int64     `json:"init_parameter"`
	AddressInfos  [2]map[string]string `json:"address_infos"`
}

type TopologyInstance struct {
	Type       string                             `json:"type"`
	Extra      map[string]string                  `json:"extra"`
	DeviceInfo map[string]DeviceRequireInfo `json:"device_need"`
}

type AddTopologyReq struct {
	Instances []TopologyInstance `json:"instances"`
	Links     []TopologyLink     `json:"links"`
}
