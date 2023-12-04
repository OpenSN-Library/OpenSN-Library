package model

type NamespaceConfig struct {
	ImageMap           map[string]string
	InterfaceAllocated []string
	ContainerEnvs      map[string]string
}

type InstanceConfig struct {
	InstanceID         string            `json:"instance_id"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	PositionChangeable bool              `json:"position_changeable"`
	Extra              map[string]string `json:"extra"`
	LinkIDs            []string
}

type LinkConfig struct {
	LinkID     string
	InstanceID [2]string
	Type       string
	Parameter  map[string]int64
}

type Namespace struct {
	Name               string
	Running            bool
	AllocatedInstances int
	NsConfig           NamespaceConfig
	InstanceConfig     []InstanceConfig
	LinkConfig         []LinkConfig
}
