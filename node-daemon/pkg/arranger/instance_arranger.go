package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"NodeDaemon/utils"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func ArrangeInstance(namespace *model.Namespace) (map[int][]model.InstanceConfig, error) {
	data.NodeMapLock.Lock()
	defer data.NodeMapLock.Unlock()
	logrus.Infof("node map: %v", data.NodeMap)
	actions := make(map[int][]model.InstanceConfig)
	left := namespace.AllocatedInstances
	ptr := 0
	for index, node := range data.NodeMap {
		allocate := utils.Min(left, node.FreeInstance)
		left -= allocate
		node.FreeInstance -= allocate

		actions[index] = namespace.InstanceConfig[ptr : ptr+allocate]
		ptr += allocate
		if left <= 0 {
			break
		}
	}
	if left > 0 {
		errMsg := fmt.Sprintf("%d instance in need, but %d left", namespace.AllocatedInstances, namespace.AllocatedInstances-left)
		err := errors.New(errMsg)
		for index, action := range actions {
			data.NodeMap[index].FreeInstance += len(action)
		}
		return nil, err
	}
	return actions, nil
}
