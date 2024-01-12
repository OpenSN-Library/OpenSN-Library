package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"container/list"
	"errors"
	"sort"

	"github.com/sirupsen/logrus"
)

type InstanceQueue []*model.InstanceConfig

func (s InstanceQueue) Len() int {
	return len(s)
}
func (s InstanceQueue) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s InstanceQueue) Less(i, j int) bool {
	return len(s[i].DeviceInfo) > len(s[j].DeviceInfo)
}

func checkNodeValid(node *model.Node, instance *model.InstanceConfig) bool {
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

func ArrangeInstance(namespace *model.Namespace) (map[int][]*model.InstanceConfig, error) {

	data.NodeMapLock.Lock()
	defer data.NodeMapLock.Unlock()
	actions := make(map[int][]*model.InstanceConfig)
	var instanceQueue InstanceQueue
	for i := range namespace.InstanceConfig {
		instanceQueue = append(instanceQueue, &namespace.InstanceConfig[i])
	}
	sort.Sort(instanceQueue)
	left := namespace.AllocatedInstances

	nodeLinkList := list.New()

	for _, v := range data.NodeMap {
		nodeLinkList.PushBack(*v)
	}

	for _, instObject := range instanceQueue {
		for ptr := nodeLinkList.Front(); ptr != nil; ptr = ptr.Next() {
			node := ptr.Value.(model.Node)
			if checkNodeValid(&node, instObject) {
				if _, ok := actions[int(node.NodeID)]; ok {
					actions[int(node.NodeID)] = append(actions[int(node.NodeID)], instObject)
				} else {
					actions[int(node.NodeID)] = []*model.InstanceConfig{instObject}
				}
				node.FreeInstance -= 1
				left -= 1
				nodeLinkList.Remove(ptr)
				nodeLinkList.PushFront(node)
				break
			}
		}

	}

	if left <= 0 {
		for ptr := nodeLinkList.Front(); ptr != nil; ptr = ptr.Next() {
			node := ptr.Value.(model.Node)
			data.NodeMap[int(node.NodeID)].FreeInstance = node.FreeInstance
			for k, v := range node.NodeLinkDeviceInfo {
				data.NodeMap[int(node.NodeID)].NodeLinkDeviceInfo[k] = v
			}
		}

	} else {
		errMsg := "Unable to arrange Instances: lack of hardware deivce"
		logrus.Error(errMsg)
		return nil, errors.New("Unable to arrange Instances: lack of hardware deivce")
	}

	return actions, nil
}
