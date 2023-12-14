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
