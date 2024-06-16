package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/synchronizer"
	"container/list"
	"fmt"

	"github.com/sirupsen/logrus"
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

func ArrangeInstances(instances []*model.Instance) error {
	nodes, err := synchronizer.GetNodeList()
	if err != nil {
		return err
	}
	nodeList := list.New()
	for _, v := range nodes {
		nodeList.PushBack(v)
	}
	for _, instance := range instances {
		success := false
		for node := nodeList.Front(); node != nil; node = node.Next() {
			nodeVal := node.Value.(*model.Node)
			if checkNodeValid(nodeVal, instance) {
				instance.NodeIndex = nodeVal.NodeIndex
				nodeVal.FreeInstance -= 1
				nodeList.Remove(node)
				nodeList.PushFront(nodeVal)
				success = true
				break
			}
		}
		if !success {
			return fmt.Errorf("unable to allocate instances")
		}
	}

	for _, v := range nodes {
		err := synchronizer.AddNode(v)
		if err != nil {
			logrus.Errorf("Update Node Info Error: %s", err.Error())
		}
	}
	return nil
}
