package arranger

import (
	"MasterNode/data"
	"MasterNode/model"
	"MasterNode/utils"
	"errors"
	"fmt"
)

func ArrangeInstance(namespace *model.Namespace) (actions map[int][]model.InstanceConfig, err error) {
	data.NodeMapLock.Lock()
	defer data.NodeMapLock.Lock()
	left := namespace.AllocatedInstances
	ptr := 0
	for index, node := range data.NodeMap {
		allocate := utils.Min(left, node.FreeInstance)
		left -= allocate
		node.FreeInstance -= allocate
		ptr += allocate
		actions[index] = namespace.InstanceConfig[ptr : ptr+allocate]
		if left <= 0 {
			break
		}
	}
	if left > 0 {
		errMsg := fmt.Sprintf("%d instance in need, but %d left", namespace.AllocatedInstances, namespace.AllocatedInstances-left)
		err = errors.New(errMsg)
		for index, action := range actions {
			data.NodeMap[index].FreeInstance += len(action)
		}
		actions = nil
	}
	return
}
