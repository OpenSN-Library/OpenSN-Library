package module

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"

	"sync"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WebShellModule struct {
	Base
}

func UpdateWebShellState(newWebshell *model.WebShellAllocRequest, oldWebshell *model.WebShellAllocRequest) error {

	if newWebshell.WebShellID != oldWebshell.WebShellID {
		if newWebshell.WebShellID != "" {
			logrus.Infof("Start Webshell: %s", newWebshell.WebShellID)
			port, err := utils.AllocTcpPort(key.SelfNode.L3AddrV4)

			if err != nil {
				return fmt.Errorf("alloc tcp port error: %s", err.Error())
			}
			logrus.Infof("Alloc Port: %d", port)
			info, err := utils.StartWebShell(
				utils.FormatIPv4(key.SelfNode.L3AddrV4),
				port,
				newWebshell.Writeable,
				newWebshell.Command,
				newWebshell.Args,
				newWebshell.ExpireMinute,
				func() {
					err := synchronizer.DelGetWebshellInfo(key.NodeIndex, newWebshell.WebShellID)
					if err != nil {
						logrus.Errorf("delete webshell info error: %s", err.Error())
					}
					err = synchronizer.DeleteWebshellRequest(key.NodeIndex, newWebshell.WebShellID)
					if err != nil {
						logrus.Errorf("delete webshell request error: %s", err.Error())
					}
				},
			)
			if err != nil {
				return fmt.Errorf("start webshell error: %s", err.Error())
			}
			err = synchronizer.UpdateGetWebshellInfo(key.NodeIndex, newWebshell.WebShellID, &model.WebShellAllocInfo{
				WebShellID: newWebshell.WebShellID,
				Addr:       info.Addr,
				Port:       info.Port,
				Pid:        info.Pid,
			})
			if err != nil {
				return fmt.Errorf("update webshell info error: %s", err.Error())
			}
		}
	}
	return nil
}

func webShellDaemonFunc(sigChan chan int, errChan chan error) {

	ctx, cancel := context.WithCancel(context.Background())
	watchChan := utils.EtcdClient.Watch(
		ctx,
		key.NodeWebshellRequestKeySelf,
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
	)
	for {

		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				cancel()
				return
			}
		case res := <-watchChan:
			for _, v := range res.Events {
				logrus.Infof("Webshell Event: %s", v.Kv.Key)
				go func(v *clientv3.Event) {
					oldWebshell := new(model.WebShellAllocRequest)
					newWebshell := new(model.WebShellAllocRequest)
					if v.PrevKv != nil && len(v.PrevKv.Value) > 0 {
						err := json.Unmarshal(v.PrevKv.Value, &oldWebshell)
						if err != nil {
							errMsg := fmt.Sprintf("Parse Link Parameter Info From %s Error: %s", string(v.Kv.Key), err.Error())
							logrus.Error(errMsg)
							return
						}
					}

					if v.Kv != nil && len(v.Kv.Value) > 0 {
						err := json.Unmarshal(v.Kv.Value, &newWebshell)
						if err != nil {
							errMsg := fmt.Sprintf("Parse Link Parameter Info From %s Error: %s", string(v.Kv.Key), err.Error())
							logrus.Error(errMsg)
							return
						}
					}

					err := UpdateWebShellState(newWebshell, oldWebshell)

					if err != nil {
						errMsg := fmt.Sprintf("Update Webshell State Error: %s", err.Error())
						logrus.Error(errMsg)
					}

				}(v)
			}
		}
	}
}

func CreateWebshellModule() *LinkParameterModule {
	return &LinkParameterModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: webShellDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "WebShell Module",
		},
	}
}
