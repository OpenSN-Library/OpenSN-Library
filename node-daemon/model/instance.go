package model

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}

type Instance struct {
	InstanceID string                       `json:"instance_id"`
	Name       string                       `json:"name"`
	Type       string                       `json:"type"`
	Image      string                       `json:"position_changeable"`
	Extra      map[string]string            `json:"extra"`
	DeviceInfo map[string]DeviceRequireInfo `json:"device_need"`
	Resource   ResourceLimit                `json:"resource"`
	NodeIndex  int                          `json:"node_id"`
	LinkIDs    []string                     `json:"link_ids"`
	Start      bool                         `json:"start"`
}

type InstanceRuntime struct {
	InstanceID  string `json:"instance_id"`
	State       string `json:"state"`
	Pid         int    `json:"pid"`
	ContainerID string `json:"container_id"`
}
