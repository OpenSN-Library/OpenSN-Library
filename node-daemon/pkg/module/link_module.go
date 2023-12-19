package module

import (
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"runtime"

	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func netLinkDaemon(requestChann chan netreq.NetLinkRequest, sigChan chan int, errChan chan error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	originNs, err := netns.Get()
	if err != nil {
		logrus.Errorf("Netlink Daemon Get Origin Netns Error: %s", err.Error())
		errChan <- err
		return
	}
	for {
		select {
		case req := <-requestChann:
			fmt.Println(req)
			netns.Set(netns.NsHandle(req.GetLinkNamespaceFd()))
			defer netns.Set(originNs)
			switch req.GetRequestType() {
			case netreq.SetLinkState:
				realReq := req.(*netreq.SetStateReq)
				fmt.Println(realReq)
			}
			err := req.ReleseFd()
			errChan <- err
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				logrus.Info("NetLink Daemon Routine Exit...")
				return
			}
		}
	}
}

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	Base
}

func linkDaemonFunc(sigChann chan int, errChann chan error) {
	netLinkReqChann := make(chan netreq.NetLinkRequest)
	netLinkSigChann := make(chan int)
	netLinkErrChann := make(chan error)
	go netLinkDaemon(netLinkReqChann, netLinkSigChann, netLinkErrChann)
	watchChan := make(chan clientv3.WatchResponse)
	go func() {
		watch := utils.EtcdClient.Watch(context.Background(), key.NodeInstanceListKeySelf)
		res := <-watch
		logrus.Infof("Etcd Instance Change Detected in Node %d", key.NodeIndex)
		watchChan <- res
	}()
	for {
		select {
		case sig := <-sigChann:
			if sig == signal.STOP_SIGNAL {
				return
			}
		case res := <-watchChan:
			fmt.Println(res)
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
			ModuleName: "LinkManage",
		},
	}
}
