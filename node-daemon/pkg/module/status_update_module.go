package module

import (
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var StatusUpdateGap = 15 * time.Second

type StatusUpdateModule struct {
	Base
}

func statusUpdateDaemonFunc(sigChan chan int, errChan chan error) {
	for {
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				logrus.Info("Status Update Module Received Stop Signal, Prepare to stop...")
				return
			}
		case <-time.After(StatusUpdateGap):
			err := utils.DoWithRetry(func() error {
				setResp := utils.RedisClient.HSet(
					context.Background(),
					key.NodeHeartBeatKey,
					strconv.Itoa(key.NodeIndex),
					strconv.FormatInt(time.Now().Unix(), 10),
				)
				if setResp.Err() != nil {
					return setResp.Err()
				}
				return nil
			}, 3)
			if err != nil {
				errMsg := fmt.Sprintf("Update Node %d Heart Beat Error %s", key.NodeIndex, err.Error())
				errChan <- err
				logrus.Error(errMsg)
				return
			}
		}
	}
}

func CreateStatusUpdateModule() *StatusUpdateModule {
	return &StatusUpdateModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: statusUpdateDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "StatusUpdate",
		},
	}
}
