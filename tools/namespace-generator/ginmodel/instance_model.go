package ginmodel

type InstanceAbstract struct {
	InstanceID  string   `json:"instance_id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	LinkIDs     []string `json:"link_ids"`
	ContainerID string   `json:"container_id"`
	Pid         int      `json:"pid"`
	Namespace   string   `json:"namespace"`
}
