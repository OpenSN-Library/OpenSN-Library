package model

type ResourceLimit struct {
	NanoCPU    int64 `json:"nano_cpu"`
	MemoryByte int64 `json:"memory_byte"`
}

type NamespaceConfig struct {
	ImageMap           map[string]string
	InterfaceAllocated []string
	ContainerEnvs      map[string]string
	ResourceLimitMap   map[string]ResourceLimit
}

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}

type InstanceConfig struct {
	InstanceID  string                       `json:"instance_id"`
	Name        string                       `json:"name"`
	Type        string                       `json:"type"`
	Image       string                       `json:"position_changeable"`
	Extra       map[string]string            `json:"extra"`
	InitLinkIDs []string                     `json:"link_ids"`
	DeviceInfo  map[string]DeviceRequireInfo `json:"device_need"`
	Resource    ResourceLimit                `json:"resource"`
}

type LinkConfig struct {
	LinkID        string               `json:"link_id"`
	EndInfos      [2]EndInfoType       `json:"init_end_infos"`
	Type          string               `json:"type"`
	InitParameter map[string]int64     `json:"init_parameter"`
	AddressInfos  [2]map[string]string `json:"address_infos"`
	LinkIndex     int                  `json:"link_index"`
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
