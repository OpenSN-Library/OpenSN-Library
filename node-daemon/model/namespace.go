package model

type NamespaceConfig struct {
	ImageMap           map[string]string
	InterfaceAllocated []string
	ContainerEnvs      map[string]string
}

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}

type InstanceConfig struct {
	InstanceID         string                       `json:"instance_id"`
	Name               string                       `json:"name"`
	Type               string                       `json:"type"`
	Image              string                       `json:"image"`
	PositionChangeable bool                         `json:"position_changeable"`
	Extra              map[string]string            `json:"extra"`
	LinkIDs            []string                     `json:"link_ids"`
	DeviceInfo         map[string]DeviceRequireInfo `json:"device_need"`
}

type LinkConfig struct {
	LinkID         string           `json:"link_id"`
	InitInstanceID [2]string        `json:"init_instance_id"`
	Type           string           `json:"type"`
	InitParameter  map[string]int64 `json:"init_parameter"`
	IPInfos        [2]IPInfoType    `json:"ip_info"`
	
}

type Namespace struct {
	Name               string           `json:"name"`
	Running            bool             `json:"running"`
	AllocatedInstances int              `json:"allocated_instances"`
	NsConfig           NamespaceConfig  `json:"ns_config"`
	InstanceAllocInfo  map[int][]string `json:"instance_alloc_info"`
	LinkAllocInfo      map[int][]string `json:"link_alloc_info"`
	InstanceConfig     []InstanceConfig `json:"instance_config"`
	LinkConfig         []LinkConfig     `json:"link_config"`
}
