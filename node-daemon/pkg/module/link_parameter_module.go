package module

import (
	"NodeDaemon/pkg/link"
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

type LinkParameterModule struct {
	Base
}

func linkParameterDaemonFunc(sigChan chan int, errChan chan error) {

	var keyLockMap sync.Map
	ctx, cancel := context.WithCancel(context.Background())
	watchChan := utils.EtcdClient.Watch(
		ctx,
		key.NodeLinkParameterKeySelf,
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
				oldParameter := make(map[string]int64)
				newParameter := make(map[string]int64)
				if v.PrevKv != nil && len(v.PrevKv.Value) > 0 {
					etcdKey = string(v.PrevKv.Key)
					err := json.Unmarshal(v.PrevKv.Value,&oldParameter)
					if err != nil {
						errMsg := fmt.Sprintf("Parse Link Parameter Info From %s Error: %s", string(v.Kv.Key), err.Error())
						logrus.Error(errMsg)
						return
					}
				}

				if v.Kv != nil && len(v.Kv.Value) > 0 {
					etcdKey = string(v.Kv.Key)
					err := json.Unmarshal(v.Kv.Value,&newParameter)
					if err != nil {
						errMsg := fmt.Sprintf("Parse Link Parameter Info From %s Error: %s", string(v.Kv.Key), err.Error())
						logrus.Error(errMsg)
						return
					}
				}

				lockAny, _ := keyLockMap.LoadOrStore(etcdKey, new(sync.Mutex))
				lock := lockAny.(*sync.Mutex)
				lock.Lock()
				defer lock.Unlock()
				linkID,_ := utils.GetEtcdLastKey(etcdKey)
				linkBase,err := synchronizer.GetLinkInfo(key.NodeIndex,linkID)
				if err != nil {
					errMsg := fmt.Sprintf("Get Link %s of Node %d Error: %s",linkID,key.NodeIndex,err.Error())
					logrus.Error(errMsg)
				}
				linkBase.Parameter = oldParameter
				linkInfo,_ := link.ParseLinkFromBase(*linkBase)
				requests,err := linkInfo.SetParameters(newParameter)
				if err != nil {
					errMsg := fmt.Sprintf("Generate Link Parameter Request for %s Error: %s",linkID,err.Error())
					logrus.Error(errMsg)
				}
				NetlinKOperatorInfo.RequestChann <- requests
			}, res.Events, 32)
			wg.Wait()
		}
	}
}

func CreateLinkParameterModule() *LinkParameterModule {
	return &LinkParameterModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: linkParameterDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "Link Parameter Module",
		},
	}
}
