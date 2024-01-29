package module

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const OperatorNum = 4

var NetlinKOperatorInfo = model.NetlinkOperatorInfo{
	RequestChann: make(chan []netreq.NetLinkRequest, OperatorNum),
}

type NetlinkOperatorModule struct {
	Base
}

func CreateNetlinkOperatorModule() *ConfigWriteModule {
	return &ConfigWriteModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			wg:         new(sync.WaitGroup),
			daemonFunc: netlinkOperatorDaemon,
			running:    false,
			ModuleName: "Netlink Operator Module",
		},
	}
}

func netlinkOperatorDaemon(sigChan chan int, errChan chan error) {
	netLinkSigChan := make(chan int)

	for i := 0; i < OperatorNum; i++ {
		go netLinkOperator(NetlinKOperatorInfo.RequestChann, netLinkSigChan, i)
	}

	select {
	case res := <-sigChan:
		if res == signal.STOP_SIGNAL {
			netLinkSigChan <- res
			return
		}
	}
}

func setLinkState(link netlink.Link, req netreq.NetLinkRequest) error {
	realReq := req.(*netreq.SetStateReq)
	if realReq.Enable {
		err := netlink.LinkSetUp(link)
		if err != nil {
			logrus.Errorf("Set Link %s Up error: %s", realReq.LinkName, err.Error())
			return err
		}
	} else {
		err := netlink.LinkSetDown(link)
		if err != nil {
			logrus.Errorf("Set Link %s Down error: %s", realReq.LinkName, err.Error())
			return err
		}
	}
	return nil
}

func setV4Addr(link netlink.Link, req netreq.NetLinkRequest) error {
	realReq := req.(*netreq.SetV4AddrReq)
	addr := strings.Split(realReq.V4Addr, "/")
	if len(addr) < 2 {
		err := fmt.Errorf("invalid ipv4 addr %s", realReq.V4Addr)
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}
	ip := net.ParseIP(addr[0])
	prefixLen, err := strconv.Atoi(addr[1])
	if err != nil {
		err = fmt.Errorf("invalid ipv4 addr prefix length %s", err.Error())
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}

	netlinkAddr := netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: utils.CreateV4InetMask(prefixLen),
		},
	}
	err = netlink.AddrAdd(link, &netlinkAddr)
	if err != nil {
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}
	return nil
}

func setV6Addr(link netlink.Link, req netreq.NetLinkRequest) error {
	realReq := req.(*netreq.SetV6AddrReq)
	addr := strings.Split(realReq.V6Addr, "/")
	if len(addr) < 2 {
		err := fmt.Errorf("invalid ipv6 addr %s", realReq.V6Addr)
		logrus.Errorf("Set Link %s IPv6 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}
	ip := net.ParseIP(addr[0])
	prefixLen, err := strconv.Atoi(addr[1])
	if err != nil {
		err = fmt.Errorf("invalid ipv6 addr %s: %s", realReq.V6Addr, err.Error())
		logrus.Errorf("Set Link %s IPv6 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}

	netlinkAddr := netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: utils.CreateV6InetMask(prefixLen),
		},
	}
	err = netlink.AddrAdd(link, &netlinkAddr)
	if err != nil {
		logrus.Errorf("Set Link %s IPv6 Addr error: %s", realReq.LinkName, err.Error())
		return err
	}
	return nil
}

func setLinkNetns(link netlink.Link, req netreq.NetLinkRequest) error {
	realReq := req.(*netreq.SetNetNsReq)
	err := netlink.LinkSetNsPid(link, realReq.TargetNamespacePid)
	if err != nil {
		logrus.Errorf("Set Link %s Netns to Pid %derror: %s", realReq.LinkName, realReq.NamespacePid, err.Error())
		return err
	}
	return nil
}

func addQdisc(link netlink.Link, req netreq.NetLinkRequest) error {
	realReq := req.(*netreq.SetQdiscReq)
	realReq.QdiscInfo.Attrs().LinkIndex = link.Attrs().Index
	err := netlink.QdiscReplace(realReq.QdiscInfo)
	if err != nil {
		logrus.Errorf("Netlink operation Error, Type %d, error: %s", req.GetRequestType(), err.Error())
		return err
	}
	return nil
}

func deleteLink(link netlink.Link, req netreq.NetLinkRequest) error {
	err := netlink.LinkDel(link)
	if err != nil {
		logrus.Errorf("Netlink operation Error, Type %d, error: %s", req.GetRequestType(), err.Error())
		return err
	}
	return nil
}

var NetReqFuncMap = map[int]func(link netlink.Link, req netreq.NetLinkRequest) error{
	netreq.SetLinkState: setLinkState,
	netreq.SetV4Addr:    setV4Addr,
	netreq.SetV6Addr:    setV6Addr,
	netreq.SetNetNs:     setLinkNetns,
	netreq.SetQdisc:     addQdisc,
	netreq.DeleteLink:   deleteLink,
}

func netLinkOperator(requestsChan chan []netreq.NetLinkRequest, sigChan chan int, index int) {
	runtime.LockOSThread()
	originNs, err := netns.Get()
	if err != nil {
		logrus.Errorf("Netlink Daemon Get Origin Netns Error: %s", err.Error())
		return
	}
	for {
		select {
		case reqs := <-requestsChan:
			for _, req := range reqs {
				var linkNsFd netns.NsHandle
				if req.GetLinkNamespacePid() != os.Getpid() {
					linkNsFd, err = netns.GetFromPid(req.GetLinkNamespacePid())
					if err != nil {
						logrus.Errorf("Get net namespace from pid %d error: %s", req.GetLinkNamespacePid(), err.Error())
					}
					err = netns.Set(linkNsFd)
					if err != nil {
						logrus.Errorf("Set net namespace error: %s", err.Error())
					}
				}
				link, err := netlink.LinkByName(req.GetLinkName())
				if err != nil {
					logrus.Errorf("Get link from Name %s in %v ,pid %d error: %s", req.GetLinkName(), linkNsFd, req.GetLinkNamespacePid(), err.Error())
				} else if operator, ok := NetReqFuncMap[req.GetRequestType()]; ok {
					err := operator(link, req)
					if err != nil {
						logrus.Errorf("NetLink Operator Error, Type: %d, Error:%s", req.GetRequestType(), err.Error())
					}
				} else {
					logrus.Errorf("Unsupport Request Type: %d", req.GetRequestType())
				}
				if req.GetLinkNamespacePid() != os.Getpid() {
					err = linkNsFd.Close()
					if err != nil {
						logrus.Errorf("Close Open Pid Netns Error: %s", err.Error())
					}
					err = netns.Set(originNs)
					if err != nil {
						logrus.Errorf("Set Back to Origin Netns Error: %s", err.Error())
					}
				}
			}
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				sigChan <- sig
				logrus.Infof("NetLink Daemon Routine %d Exit...", index)
				return
			}
		}
	}
}
