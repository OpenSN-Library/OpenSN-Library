package arranger

import (
	"NodeDaemon/model"
	"sync"

	"github.com/sirupsen/logrus"
)

var nextIndex int = 1
var nextIndexLock *sync.Mutex = new(sync.Mutex)

func ArrangeLink(namespace *model.Namespace, instanceTarget map[int][]*model.InstanceConfig) (map[int][]*model.LinkConfig, error) {
	actions := make(map[int][]*model.LinkConfig)
	nodeIndexSet := make(map[string]int)
	for index, v := range instanceTarget {
		for _, instance := range v {
			nodeIndexSet[instance.InstanceID] = index
		}
	}
	for link_index, linkInfo := range namespace.LinkConfig {
		nextIndexLock.Lock()
		namespace.LinkConfig[link_index].LinkIndex = nextIndex
		nextIndex = (nextIndex + 1) % (1 << 24)
		nextIndexLock.Unlock()
		if linkInfo.InitEndInfos[0].InstanceID == "" {
			targetIndex1 := nodeIndexSet[linkInfo.InitEndInfos[1].InstanceID]
			namespace.LinkConfig[link_index].InitEndInfos[1].EndNodeIndex = targetIndex1
			actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])

		} else if linkInfo.InitEndInfos[1].InstanceID == "" {
			targetIndex0 := nodeIndexSet[linkInfo.InitEndInfos[0].InstanceID]
			namespace.LinkConfig[link_index].InitEndInfos[0].EndNodeIndex = targetIndex0
			actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])

		} else {
			targetIndex0 := nodeIndexSet[linkInfo.InitEndInfos[0].InstanceID]
			targetIndex1 := nodeIndexSet[linkInfo.InitEndInfos[1].InstanceID]
			namespace.LinkConfig[link_index].InitEndInfos[0].EndNodeIndex = targetIndex0
			namespace.LinkConfig[link_index].InitEndInfos[1].EndNodeIndex = targetIndex1

			logrus.Infof("Add Link Between %s and %s", linkInfo.InitEndInfos[0].InstanceID, linkInfo.InitEndInfos[1].InstanceID)
			if targetIndex0 == targetIndex1 {
				actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])
			} else {
				actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])
				actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])
			}
		}
	}
	return actions, nil
}
