package ginmodel

import "NodeDaemon/model"

type SingleInstanceRequest struct {
	InstanceID string `json:"instance_id"`
	NodeIndex  int    `json:"node_index"`
}

type GetInstanceListRequest struct {
	KeyWord   string `json:"key_word"`
	PageSize  int    `json:"page_size"`
	PageIndex int    `json:"page_index"`
}

type InstanceAbstract struct {
	InstanceID string            `json:"instance_id"`
	Start      bool              `json:"start"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	NodeIndex  int               `json:"node_index"`
	Extra      map[string]string `json:"extra"`
}

type AddInstanceRequest struct {
	Type       string                             `json:"type"`
	Extra      map[string]string                  `json:"extra"`
	DeviceInfo map[string]model.DeviceRequireInfo `json:"device_need"`
}

type InstanceInfo struct {
	InstanceID    string                          `json:"instance_id"`
	Name          string                          `json:"name"`
	Type          string                          `json:"type"`
	Image         string                          `json:"image"`
	Start         bool                            `json:"start"`
	Extra         map[string]string               `json:"extra"`
	ResourceLimit model.ResourceLimit             `json:"resource_limit"`
	Connections   map[string]model.ConnectionInfo `json:"connections"`
	NodeIndex     int                             `json:"node_index"`
}
