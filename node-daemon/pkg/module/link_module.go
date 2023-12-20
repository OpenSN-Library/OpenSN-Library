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
	"github.com/vishvananda/netlink"
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
			var opErr error
			linkNsFd, err := netns.GetFromPid(req.GetLinkNamespacePid())
			if err != nil {
				errChan <- err
				continue
			}
			err = netns.Set(linkNsFd)
			if err != nil {
				errChan <- err
				continue
			}
			link, err := netlink.LinkByIndex(req.GetLinkIndex())
			if err != nil {
				errChan <- err
				continue
			}
			switch req.GetRequestType() {
			case netreq.SetLinkState:
				realReq := req.(*netreq.SetStateReq)
				if realReq.Enable {
					opErr = netlink.LinkSetUp(link)
				} else {
					opErr = netlink.LinkSetDown(link)
				}
			case netreq.SetV4Addr:
				realReq := req.(*netreq.SetV4AddrReq)
				addr := netlink.Addr{
					IPNet: utils.CreateV4Inet(realReq.V4Addr, realReq.PrefixLen),
				}
				opErr = netlink.AddrAdd(link, &addr)
			case netreq.SetV6Addr:
				realReq := req.(*netreq.SetV6AddrReq)
				addr := netlink.Addr{
					IPNet: utils.CreateV6Inet(realReq.V6Addr, realReq.PrefixLen),
				}
				opErr = netlink.AddrAdd(link, &addr)
			case netreq.SetNetNs:
				realReq := req.(*netreq.SetNetNsReq)
				opErr = netlink.LinkSetNsPid(link, realReq.TargetNamespacePid)
			case netreq.SetQdisc:
				realReq := req.(*netreq.SetQdiscReq)
				opErr = netlink.QdiscReplace(realReq.QdiscInfo)
			default:
				logrus.Errorf("Unsupport Request Type: %d", req.GetRequestType())
			}
			err = netns.Set(originNs)
			if err != nil {
				errChan <- err
				continue
			}
			err = linkNsFd.Close()
			if err != nil {
				errChan <- err
				continue
			}
			errChan <- opErr
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
