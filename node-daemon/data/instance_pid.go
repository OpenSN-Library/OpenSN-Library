package data

import (
	"sync"
	"time"
)

type InstancePidPair struct {
	InstanceID string
	Pid        int
}

var ContainerInstancePidMap = new(sync.Map)

func WatchInstancePid(instanceID string) int {
	val, _ := ContainerInstancePidMap.LoadOrStore(instanceID, make(chan int))
	ch := val.(chan int)
	res := <-ch
	select {
	case _, received := <-ch:
		if received {
			ch <- res
		}
	default:
		ch <- res
	}
	return res
}

func DeleteInstancePid(instanceID string, timeout time.Duration) {
	ContainerInstancePidMap.Delete(instanceID)
}

func SetInstancePid(instanceID string, pid int) {
	val, _ := ContainerInstancePidMap.LoadOrStore(instanceID, make(chan int))
	ch := val.(chan int)

	select {
	case <-ch:
	default:
	}
	ch <- int(pid)
}

func GetAllPids() []InstancePidPair {
	pairList := []InstancePidPair{}
	ContainerInstancePidMap.Range(func(key, value any) bool {
		id := key.(string)
		ch := value.(chan int)
		select {
		case pid := <-ch:
			pairList = append(pairList, InstancePidPair{
				InstanceID: id,
				Pid:        pid,
			})
			ch <- pid
		default:
		}
		return true
	})
	return pairList
}
