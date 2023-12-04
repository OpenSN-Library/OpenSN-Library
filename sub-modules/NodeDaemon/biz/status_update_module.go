package biz

import (
	"NodeDaemon/config"
	"NodeDaemon/data"
	"NodeDaemon/utils"
	"NodeDaemon/utils/tools"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var StatusUpdateGap = 15 * time.Second

type StatusUpdateModule struct {
	ModuleBase
}

func statusUpdateDaemonFunc(sigChan chan int, errChan chan error) {
	for {
		select {
		case sig := <-sigChan:
			if sig == config.STOP_SIGNAL {
				return
			}
		case <-time.After(StatusUpdateGap):
			err := tools.DoWithRetry(func() error {
				setResp := utils.RedisClient.HSet(
					context.Background(),
					config.NodeHeartBeatKey,
					strconv.Itoa(data.NodeIndex),
					strconv.FormatInt(time.Now().Unix(), 10),
				)
				if setResp.Err() != nil {
					return setResp.Err()
				}
				return nil
			}, 3)
			if err != nil {
				errMsg := fmt.Sprintf("Update Node %d Heart Beat Error %s", data.NodeIndex, err.Error())
				errChan <- err
				logrus.Error(errMsg)
				return
			}
		}
	}
}

func CreateStatusUpdateModule() *StatusUpdateModule {
	return &StatusUpdateModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: statusUpdateDaemonFunc,
			wg:         new(sync.WaitGroup),
		},
	}
}
