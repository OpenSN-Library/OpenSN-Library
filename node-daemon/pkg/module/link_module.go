package module

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/link"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"

	"sync"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func UpdateLinkState(newLink model.Link) error {
	newState := true
	for _, endInfo := range newLink.GetEndInfos() {
		if endInfo.EndNodeIndex == key.NodeIndex && endInfo.InstancePid == 0 {
			newState = false
		}
	}

	if newState != newLink.IsEnabled() {

		if newState {
			requests, err := newLink.Enable()
			if err != nil {
				return fmt.Errorf("generate enable requests for link %s error: %s", newLink.GetLinkID(), err.Error())
			}
			NetlinKOperatorInfo.RequestChann <- requests
		} else {
			requests, err := newLink.Disable()
			if err != nil {
				return fmt.Errorf("generate disable requests for link %s error: %s", newLink.GetLinkID(), err.Error())
			}
			NetlinKOperatorInfo.RequestChann <- requests
		}

		synchronizer.UpdateLinkInfo(
			key.NodeIndex,
			newLink.GetLinkID(),
			func(lb *model.LinkBase) error {
				lb.Enabled = newState
				return nil
			},
		)
	}
	return nil
}

type LinkModule struct {
	Base
}

func linkDaemonFunc(sigChan chan int, errChan chan error) {

	var keyLockMap sync.Map
	ctx, cancel := context.WithCancel(context.Background())
	watchChan := utils.EtcdClient.Watch(
		ctx,
		key.NodeLinkListKeySelf,
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
			wg := utils.ForEachWithThreadPool[*clientv3.Event](func(v *clientv3.Event) {
				etcdKey := ""
				oldLink, _ := link.ParseLinkFromBase(model.LinkBase{})
				newLink, _ := link.ParseLinkFromBase(model.LinkBase{})
				var err error
				if v.PrevKv != nil && len(v.PrevKv.Value) > 0 {
					etcdKey = string(v.PrevKv.Key)
					oldLink, err = link.ParseLinkFromBytes(v.PrevKv.Value)
					if err != nil {
						errMsg := fmt.Sprintf("Parse Link Info From %s Error: %s", string(v.Kv.Key), err.Error())
						logrus.Error(errMsg)
						return
					}
				}

				if v.Kv != nil && len(v.Kv.Value) > 0 {
					etcdKey = string(v.Kv.Key)
					newLink, err = link.ParseLinkFromBytes(v.Kv.Value)
					if err != nil {
						errMsg := fmt.Sprintf("Parse Link Info From %s Error: %s", string(v.Kv.Key), err.Error())
						logrus.Error(errMsg)
						return
					}
				}

				lockAny, _ := keyLockMap.LoadOrStore(etcdKey, new(sync.Mutex))
				lock := lockAny.(*sync.Mutex)
				lock.Lock()
				defer lock.Unlock()

				stateChange := false
				logrus.Debugf("old link is %v, new link is %v", oldLink, newLink)
				for endIndex, endInfo := range oldLink.GetEndInfos() {
					if newLink.GetEndInfos()[endIndex].EndNodeIndex == key.NodeIndex &&
						newLink.GetEndInfos()[endIndex].InstancePid != endInfo.InstancePid {
						stateChange = true
					}
				}

				if stateChange {
					logrus.Debugf("State Change of %s Detectd.", etcdKey)
					err := UpdateLinkState(newLink)
					if err != nil {
						errMsg := fmt.Sprintf("Do Link State Change of %s Error: %s", etcdKey, err.Error())
						logrus.Error(errMsg)
					}
					return
				}

				if newLink.IsEnabled() {
					requests, err := newLink.SetParameters(newLink.GetLinkBasePtr().Parameter)
					if err != nil {
						errMsg := fmt.Sprintf("Generate Parameter Update Requests for Link %s Error: %s", etcdKey, err.Error())
						logrus.Error(errMsg)
						return
					}
					NetlinKOperatorInfo.RequestChann <- requests
				}
			}, res.Events, 32)
			wg.Wait()
		}
	}
}

func CreateLinkModuleTask() *LinkModule {
	return &LinkModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: linkDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "Link Manager Module",
		},
	}
}
