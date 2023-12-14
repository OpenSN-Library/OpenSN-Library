package ginmodel

type NsReqConfig struct {
	ImageMap      map[string]string `json:"image_map"`
	ContainerEnvs map[string]string `json:"container_envs"`
}

type InstanceReqConfig struct {
	Type               string            `json:"type"`
	PositionChangeable bool              `json:"position_changeable"`
	Extra              map[string]string `json:"extra"`
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

type NamespaceInfoData struct {
	Name string `json:"name"`
}
