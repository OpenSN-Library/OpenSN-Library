package data

import (
	"sync"
	"time"
)

type InstancePidPair struct {
	InstanceID string
	Pid        int
}

const checkGap = 100 * time.Millisecond

var containerInstancePidMap = make(map[string]int)
var dataPoolLock = new(sync.RWMutex)

func WatchInstancePid(instanceID string) int {
	for {
		dataPoolLock.RLock()
		pid, ok := containerInstancePidMap[instanceID]
		dataPoolLock.RUnlock()
		if ok {
			return pid
		}
		time.Sleep(checkGap)
	}
}

func TryGetInstancePid(instanceID string) (int, bool) {
	dataPoolLock.RLock()
	pid, ok := containerInstancePidMap[instanceID]
	dataPoolLock.RUnlock()
	return pid, ok
}

func DeleteInstancePid(instanceID string) {
	dataPoolLock.Lock()
	delete(containerInstancePidMap, instanceID)
	dataPoolLock.Unlock()
}

func SetInstancePid(instanceID string, pid int) {
	dataPoolLock.Lock()
	containerInstancePidMap[instanceID] = pid
	dataPoolLock.Unlock()
}
