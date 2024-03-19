package module

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/link"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"

	"sync"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func UpdateLinkState(newLink model.Link, oldLink model.Link) error {

	var err error
	isEnable := oldLink.IsEnabled() || newLink.IsEnabled()
	isCreated := oldLink.IsCreated() || newLink.IsCreated()

	shouldEnable := newLink.GetLinkBasePtr().Enable
	shouldCreate := newLink.GetLinkID() != ""

	if isCreated != shouldCreate {
		if shouldCreate {
			err = newLink.Create()
		} else {
			err = oldLink.Destroy()
		}
	}

	if err != nil {
		return fmt.Errorf("update create state error: %s", err.Error())
	}

	if shouldEnable != isEnable {
		if shouldEnable {
			err = newLink.Enable()
		} else {
			err = newLink.Disable()
		}
	}
	if err != nil {
		return fmt.Errorf("update running state error: %s", err.Error())
	}
	return nil

}

type LinkModule struct {
	Base
}

func linkDaemonFunc(sigChan chan int, errChan chan error) {

	// var keyLockMap sync.Map
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
			for _, v := range res.Events {
				go func(v *clientv3.Event) {
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
					logrus.Debugf("Link %s Update Detected From %v to %v", etcdKey, oldLink, newLink)

					// lockAny, _ := keyLockMap.LoadOrStore(etcdKey, new(sync.Mutex))
					// lock := lockAny.(*sync.Mutex)
					// lock.Lock()
					// defer lock.Unlock()

					err = UpdateLinkState(newLink, oldLink)
					if err != nil {
						errMsg := fmt.Sprintf("Do Link State Change of %s Error: %s", etcdKey, err.Error())
						logrus.Error(errMsg)
					} else {
						logrus.Infof("Do Link State Change of %s Success", etcdKey)
					}
				}(v)

			}
		}
	}
}

func CreateLinkManagerModule() *LinkModule {
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
