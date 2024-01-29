package arranger

import (
	"NodeDaemon/model"
)

func checkNodeValid(node *model.Node, instance *model.Instance) bool {
	if node.FreeInstance <= 0 {
		return false
	}
	for k, v := range instance.DeviceInfo {
		nodeDev, ok := node.NodeLinkDeviceInfo[k]
		if !ok {
			return false
		}
		if nodeDev < v.NeedNum {
			return false
		}
	}
	for k, v := range instance.DeviceInfo {
		if v.IsMutex {
			node.NodeLinkDeviceInfo[k] = node.NodeLinkDeviceInfo[k] - v.NeedNum
		}
	}

	return true
}

func ArrangeInstances(instance []*model.Instance) error {

	for _, instance := range instance {
		instance.NodeIndex = 0
	}
	return nil
}
