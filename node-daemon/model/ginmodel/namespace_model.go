package ginmodel

type ResourceLimit struct {
	NanoCPU    string `json:"nano_cpu"`
	MemoryByte string `json:"memory_byte"`
}

type NsReqConfig struct {
	ImageMap      map[string]string        `json:"image_map"`
	ContainerEnvs map[string]string        `json:"container_envs"`
	ResourceMap   map[string]ResourceLimit `json:"resource_map"`
}

type InstanceReqConfig struct {
	Type  string            `json:"type"`
	Extra map[string]string `json:"extra"`
}

type LinkReqConfig struct {
	Type          string           `json:"type"`
	InstanceIndex [2]int           `json:"instance_index"`
	Parameter     map[string]int64 `json:"parameter"`
}

type CreateNamespaceReq struct {
	Name        string              `json:"name"`
	NsConfig    NsReqConfig         `json:"ns_config"`
	InstConfigs []InstanceReqConfig `json:"inst_config"`
	LinkConfigs []LinkReqConfig     `json:"link_config"`
}

type UpdateNamespaceReq struct {
	NsConfig    NsReqConfig         `json:"ns_config"`
	InstConfigs []InstanceReqConfig `json:"inst_config"`
	LinkConfigs []LinkReqConfig     `json:"link_config"`
}

type NsInstanceData struct {
}

type NsLinkData struct {
}

type NamespaceInfoData struct {
	Name              string             `json:"name"`
	Running           bool               `json:"running"`
	InstanceAllocInfo map[int][]string   `json:"instance_alloc_info"`
	LinkAllocInfo     map[int][]string     `json:"link_alloc_info"`
	InstanceInfos     []InstanceAbstract `json:"instance_infos"`
	LinkInfos         []LinkAbstract     `json:"link_infos"`
}

type NamespaceAbstract struct {
	Name           string `json:"name"`
	Running        bool   `json:"running"`
	InstanceNum    int    `json:"instance_num"`
	LinkNum        int    `json:"link_num"`
	AllocNodeIndex []int  `json:"alloc_node_index"`
}
