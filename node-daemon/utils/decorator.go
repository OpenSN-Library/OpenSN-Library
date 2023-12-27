package utils

import (
	"errors"
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

func MapKeys[K comparable](rawMap map[K]any) []K {
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
