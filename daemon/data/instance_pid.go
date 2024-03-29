package data

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type InstancePidPair struct {
	InstanceID string
	Pid        int
}

var ContainerInstancePidMap = new(sync.Map)

func WatchInstancePid(instanceID string) int {
	val, _ := ContainerInstancePidMap.LoadOrStore(instanceID, make(chan int,2))
	ch := val.(chan int)
	res := <-ch
	ch <- res
	return res
}

func TryGetInstancePid(instanceID string) (int, bool) {
	val, ok := ContainerInstancePidMap.Load(instanceID)
	if !ok {
		return 0, false
	}
	ch := val.(chan int)
	select {
	case pid := <-ch:
		ch <- pid
		return pid, true
	default:
		return 0, false
	}
}

func DeleteInstancePid(instanceID string, timeout time.Duration) {
	ContainerInstancePidMap.Delete(instanceID)
}

func SetInstancePid(instanceID string, pid int) {
	val, _ := ContainerInstancePidMap.LoadOrStore(instanceID, make(chan int,2))
	ch := val.(chan int)

	select {
	case <-ch:
	default:
	}
	select {
	case ch <- pid:
	default:
		<-ch
		ch <- pid
		logrus.Warnf("Blocked Write of Instance %s Pid", instanceID)
	}
}
