package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
)

func ArrangeLink(namespace *model.Namespace) (map[int][]*model.LinkConfig, error) {
	actions := make(map[int][]*model.LinkConfig)
	for link_index, linkInfo := range namespace.LinkConfig {
		if linkInfo.InitInstanceID[0] == "" {
			instanceInfo1 := data.InstanceMap[linkInfo.InitInstanceID[0]]
			actions[int(instanceInfo1.NodeID)] = append(actions[int(instanceInfo1.NodeID)], &namespace.LinkConfig[link_index])
		} else if linkInfo.InitInstanceID[1] == "" {
			instanceInfo2 := data.InstanceMap[linkInfo.InitInstanceID[1]]
			actions[int(instanceInfo2.NodeID)] = append(actions[int(instanceInfo2.NodeID)], &namespace.LinkConfig[link_index])
		} else {
			instanceInfo1 := data.InstanceMap[linkInfo.InitInstanceID[0]]
			instanceInfo2 := data.InstanceMap[linkInfo.InitInstanceID[1]]
			if instanceInfo1.NodeID == instanceInfo2.NodeID {
				actions[int(instanceInfo1.NodeID)] = append(actions[int(instanceInfo1.NodeID)], &namespace.LinkConfig[link_index])
			} else {
				actions[int(instanceInfo1.NodeID)] = append(actions[int(instanceInfo1.NodeID)], &namespace.LinkConfig[link_index])
				actions[int(instanceInfo2.NodeID)] = append(actions[int(instanceInfo2.NodeID)], &namespace.LinkConfig[link_index])
			}
		}
	}
	return actions, nil
}
