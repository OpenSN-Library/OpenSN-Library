package model

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type Instance struct {
	Config      InstanceConfig `json:"config"`
	ContainerID string         `json:"container_id"`
	Pid         int            `json:"pid"`
	State       string         `json:"state"`
	NodeID      uint32         `json:"node_id"`
	Namespace   string         `json:"namespace"`
	LinkIDs     []string       `json:"link_ids"`
}
