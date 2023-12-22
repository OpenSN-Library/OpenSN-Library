package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"fmt"
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

func ArrangeInstance(namespace *model.Namespace) (map[int][]*model.InstanceConfig, error) {

	data.NodeMapLock.Lock()
	defer data.NodeMapLock.Unlock()
	var allocedDev = map[int]map[string]int{}
	actions := make(map[int][]*model.InstanceConfig)
	var instanceQueue InstanceQueue
	for _, v := range namespace.InstanceConfig {
		instanceQueue = append(instanceQueue, &v)
	}
	sort.Sort(instanceQueue)
	left := namespace.AllocatedInstances
	for index, node := range data.NodeMap {
		for inst_index, inst := range instanceQueue {
			valid := true
			if inst == nil {
				continue
			}
			for devType, devInfo := range inst.DeviceInfo {
				if _, ok := node.NodeLinkDeviceInfo[devType]; !ok {
					valid = false
					break
				}
				if devInfo.IsMutex {
					if node.NodeLinkDeviceInfo[devType] < devInfo.NeedNum {
						valid = false
					} else {
						if allocedDev[index] == nil {
							allocedDev[index] = map[string]int{
								devType: devInfo.NeedNum,
							}
						} else {
							allocedDev[index][devType] += devInfo.NeedNum
						}
						node.NodeLinkDeviceInfo[devType] -= devInfo.NeedNum
					}
				} else {
					if node.NodeLinkDeviceInfo[devType] < devInfo.NeedNum {
						valid = false
					}
				}
			}

			if valid {
				actions[index] = append(actions[index], inst)
				instanceQueue[inst_index] = nil
				left -= 1
			}

		}
	}

	if left > 0 {
		var leftIdArray []string
		for _, v := range instanceQueue {
			if v != nil {
				leftIdArray = append(leftIdArray, v.InstanceID)
			}
		}
		err := fmt.Errorf("%v cannot be arrange",leftIdArray)
		logrus.Errorf("Unable to arrange Instances: %s", err.Error())

		for node_index, allocMap := range allocedDev {
			for devType, num := range allocMap {
				data.NodeMap[node_index].NodeLinkDeviceInfo[devType] += num
			}
		}
		return nil,err
	}

	return actions, nil
}
