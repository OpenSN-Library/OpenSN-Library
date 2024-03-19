package model

import (
	"NodeDaemon/utils"
	"context"
	"fmt"
)

type DeviceRequireInfo struct {
	DevName string `json:"dev_name"`
	NeedNum int    `json:"need_num"`
	IsMutex bool   `json:"is_mutex"`
}

type ConnectionInfo struct {
	LinkID       string `json:"link_id"`
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	EndNodeIndex int    `json:"end_node_index"`
}

type Instance struct {
	InstanceID  string                       `json:"instance_id"`
	Name        string                       `json:"name"`
	Type        string                       `json:"type"`
	Image       string                       `json:"position_changeable"`
	Extra       map[string]string            `json:"extra"`
	DeviceInfo  map[string]DeviceRequireInfo `json:"device_need"`
	Resource    ResourceLimit                `json:"resource"`
	NodeIndex   int                          `json:"node_index"`
	Connections map[string]ConnectionInfo    `json:"connections"`
	Start       bool                         `json:"start"`
}

func (i *Instance) IsCreated() bool {
	_, err := utils.DockerClient.ContainerInspect(context.Background(), fmt.Sprintf(
		"%s_%s", i.Type, i.InstanceID,
	))
	if err != nil {
		return false
	}
	return true
}

func (i *Instance) IsRunning() bool {
	info, err := utils.DockerClient.ContainerInspect(context.Background(), fmt.Sprintf(
		"%s_%s", i.Type, i.InstanceID,
	))
	if err != nil {
		return false
	}
	return info.State.Running
}
