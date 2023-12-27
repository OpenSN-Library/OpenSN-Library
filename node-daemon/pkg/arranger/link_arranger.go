package arranger

import (
	"NodeDaemon/model"

	"github.com/sirupsen/logrus"
)

func ArrangeLink(namespace *model.Namespace, instanceTarget map[int][]*model.InstanceConfig) (map[int][]*model.LinkConfig, error) {
	actions := make(map[int][]*model.LinkConfig)
	nodeIndexSet := make(map[string]int)
	for index, v := range instanceTarget {
		for _, instance := range v {
			nodeIndexSet[instance.InstanceID] = index
		}
	}
	for link_index, linkInfo := range namespace.LinkConfig {
		if linkInfo.InitInstanceID[0] == "" {
			targetIndex1 := nodeIndexSet[linkInfo.InitInstanceID[0]]
			actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])
		} else if linkInfo.InitInstanceID[1] == "" {
			targetIndex2 := nodeIndexSet[linkInfo.InitInstanceID[1]]
			actions[targetIndex2] = append(actions[targetIndex2], &namespace.LinkConfig[link_index])
		} else {
			targetIndex1 := nodeIndexSet[linkInfo.InitInstanceID[0]]
			targetIndex2 := nodeIndexSet[linkInfo.InitInstanceID[1]]
			logrus.Infof("Add Link Between %s and %s", linkInfo.InitInstanceID[0], linkInfo.InitInstanceID[1])
			if targetIndex1 == targetIndex2 {
				actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])
			} else {
				actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])
				actions[targetIndex2] = append(actions[targetIndex2], &namespace.LinkConfig[link_index])
			}
		}
	}
	return actions, nil
}
