package utils

import (
	"errors"
	"sync"
	"time"
)

func DoWithRetry(thing func() error, maxTime int) error {
	var err error
	for i := 0; i < maxTime; i++ {
		err = thing()
		if err == nil {
			break
		}
	}
	return err
}

func WaitSuccess(call func() bool, timeout time.Duration, checkGap time.Duration) error {
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			return errors.New("wait timeout")
		case <-time.After(checkGap):
			if call() {
				return nil
			}
		}
	}
}

func SliceMap[I, O any](call func(i I) O, slice []I) []O {
	res := make([]O, len(slice))
	for i, v := range slice {
		res[i] = call(v)
	}
	return res
}

func MapKeys[K comparable,V any](rawMap map[K]V) []K {
	index := 0
	res := make([]K, len(rawMap))
	for k := range rawMap {
		res[index] = k
		index += 1
	}
	return res
}

func Spin(check func() bool, gap time.Duration) {
	for {
		if check() {
			return
		}
		time.Sleep(gap)
	}
}

// pass by value, cannot have side-effect
func ForEachWithThreadPool[T any](callable func(v T), array []T, maxRoutine int) *sync.WaitGroup {
	chanBuf := make(chan bool, maxRoutine)
	wg := new(sync.WaitGroup)
	for _, v := range array {
		chanBuf <- true
		wg.Add(1)
		go func(v T) {
			callable(v)
			<-chanBuf
			wg.Done()
		}(v)
	}
	return wg
}

// Pass by value, cannot have side-effect
func ForEachUtilAllComplete[T any](callable func(v T) (bool, error), array []T) error {
	var queue []T
	var finalErr error
	for _, v := range array {
		queue = append(queue, v)
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		success, err := callable(v)
		if err != nil {
			finalErr = err
		}
		if !success {
			queue = append(queue, v)
		}
	}
	return finalErr
}

func ForEachUtilAllCompleteWithThreadPool[T any](callable func(v T) bool, array []T, maxRoutine int) *sync.WaitGroup {
	var queue []T
	chanBuf := make(chan bool, maxRoutine)
	lock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	queue = append(queue, array...)

	for len(queue) > 0 {
		v := queue[0]
		lock.Lock()
		queue = queue[1:]
		lock.Unlock()
		chanBuf <- true
		wg.Add(1)
		go func(v T) {
			success := callable(v)
			if !success {
				lock.Lock()
				queue = append(queue, v)
				lock.Unlock()
			}
			<-chanBuf
			wg.Done()
		}(v)
	}
	return wg
}