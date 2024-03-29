package model

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
