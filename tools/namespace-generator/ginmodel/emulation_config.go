package ginmodel

type ResourceLimit struct {
	NanoCPU    string `json:"nano_cpu"`
	MemoryByte string `json:"memory_byte"`
}

type InstanceTypeConfig struct {
	Image         string            `json:"image"`
	Envs          map[string]string `json:"container_envs"`
	ResourceLimit ResourceLimit     `json:"resource_limit"`
}

type EmulationInfo struct {
	Running    bool                          `json:"running"`
	TypeConfig map[string]InstanceTypeConfig `json:"type_config"`
}
