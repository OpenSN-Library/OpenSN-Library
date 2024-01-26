package module

import (
	"NodeDaemon/share/dir"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ConfigWriteModule struct {
	Base
}

func CreateConfigWriteModule() *ConfigWriteModule {
	return &ConfigWriteModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			wg:         new(sync.WaitGroup),
			daemonFunc: watchConfigChange,
			running:    false,
			ModuleName: "configWatchModule",
		},
	}
}

func watchConfigChange(sigChan chan int, errChan chan error) {
	for {
		ctx, cancel := context.WithCancel(context.Background())

		watchChan := utils.EtcdClient.Watch(
			ctx,
			key.NodeInstanceConfigKeySelf,
			clientv3.WithPrefix(),
		)

		for {
			select {
			case sig := <-sigChan:
				if sig == signal.STOP_SIGNAL {
					cancel()
					return
				}
			case res := <-watchChan:
				if len(res.Events) < 1 {
					logrus.Error("Unexpected Instance Config Info List Length:", len(res.Events))
					continue
				}
				for _, event := range res.Events {
					keySplit := strings.Split(string(event.Kv.Key), "/")
					instanceID := keySplit[len(keySplit)-1]
					configPath := path.Join(dir.TopoInfoDir, instanceID)
					if len(event.Kv.Value) > 0 {
						fd, err := os.Create(configPath)
						if err != nil {
							errMsg := fmt.Sprintf("Open File %s to Write Error: %s", configPath, err.Error())
							logrus.Error(errMsg)
							continue
						}
						_, err = fd.Write(event.Kv.Value)
						if err != nil {
							errMsg := fmt.Sprintf("Write to File %s Error: %s", configPath, err.Error())
							logrus.Error(errMsg)
							continue
						}
					} else {
						os.Remove(configPath)
					}
				}
			}
		}
	}
}
