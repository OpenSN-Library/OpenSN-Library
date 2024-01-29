package ginmodel

import "NodeDaemon/model"

type ConfigEmulationReq map[string]TypeConfigReq

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
	InstanceTypeConfig map[string]model.InstanceTypeConfig `json:"instance_type_config"`
	Running            bool                                `json:"running"`
}

type EmulationConfig struct {
	InstanceTypeConfig map[string]model.InstanceTypeConfig `json:"instance_type_config"`
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
	DeviceInfo map[string]model.DeviceRequireInfo `json:"device_need"`
}

type AddTopologyReq struct {
	Instances []TopologyInstance `json:"instances"`
	Links     []TopologyLink     `json:"links"`
}
