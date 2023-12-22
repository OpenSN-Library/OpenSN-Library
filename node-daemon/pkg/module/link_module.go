package module

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/pkg/link"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitLinkData() {
	getResp, err := utils.EtcdClient.Get(
		context.Background(),
		key.NodeLinkListKeySelf,
	)
	if err != nil {
		errMsg := fmt.Sprintf("Check Node Instance List Initialized %s Error: %s", key.NodeInstancesKeySelf, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	if len(getResp.Kvs) <= 0 {
		_, err := utils.EtcdClient.Put(
			context.Background(),
			key.NodeInstanceListKeySelf,
			"[]",
		)
		if err != nil {
			errMsg := fmt.Sprintf("Init Node Instance List %s Error: %s", key.NodeInstancesKeySelf, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
	}
}

func parseLinkChange(updateIdList []string) (addList []string, delList []model.Link, err error) {
	var delIDList []string
	updateIDMap := make(map[string]bool)
	for _, v := range updateIdList {
		updateIDMap[v] = true
	}
	for k := range updateIDMap {
		if _, ok := data.LinkMap[k]; !ok {
			addList = append(addList, k)
		}
	}

	for k := range data.LinkMap {
		if ok := updateIDMap[k]; !ok {
			delIDList = append(delIDList, k)
		}
	}

	for _, v := range delIDList {
		delList = append(delList, data.LinkMap[v])
		delete(data.LinkMap, v)
	}

	if len(addList) > 0 {

		redisResponse := utils.RedisClient.HMGet(context.Background(), key.NodeLinksKeySelf, addList...)

		if redisResponse.Err() != nil {
			err = redisResponse.Err()
			logrus.Error("Get Instance Infos Error: ", err.Error())
			return
		}

		for i, v := range redisResponse.Val() {
			if v == nil {
				logrus.Error("Redis Result Empty, Redis Data May Crash, InstanceID:", addList[i])
				continue
			} else {
				newLink, err := link.ParseLink([]byte(v.(string)))
				if err != nil {
					logrus.Error("Unmarshal Json Data to Link Base Error, Redis Data May Crash: ", err.Error())
					continue
				}
				data.LinkMap[addList[i]] = newLink
			}
		}
	}
	return
}

func netLinkDaemon(requestChan chan netreq.NetLinkRequest, sigChan chan int, errChan chan error) {
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
		case req := <-requestChan:
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

func linkParameterWatcher(sigChan chan int) {

}

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	Base
}

func AddLinks(addList []string, operator *model.NetlinkOperatorInfo) error {
	for _, v := range addList {
		linkInfo := data.LinkMap[v]

		err := linkInfo.Enable(operator)
		if err != nil {
			logrus.Errorf("Enable Link %s Error: %s", linkInfo.GetLinkID(), err.Error())
		}

		err = linkInfo.Connect(operator)
		if err != nil {
			logrus.Errorf("Connect Link %s Error: %s", linkInfo.GetLinkID(), err.Error())
		}
	}
	return nil
}

func DelLinks(delList []model.Link, operator *model.NetlinkOperatorInfo) error {
	for _, v := range delList {
		if v.IsConnected() {
			err := v.Disconnect(operator)
			if err != nil {
				logrus.Errorf("Disconnect Link %s Error: %s", v.GetLinkID(), err.Error())
			}
		}
		err := v.Disable(operator)
		if err != nil {
			logrus.Errorf("Disable Link %s Error: %s", v.GetLinkID(), err.Error())
		}
		delete(data.LinkMap, v.GetLinkID())
	}
	return nil
}

func linkDaemonFunc(sigChan chan int, errChan chan error) {
	InitLinkData()
	netOpInfo := model.NetlinkOperatorInfo{
		RequestChann: make(chan netreq.NetLinkRequest),
		ErrChan:      make(chan error),
	}
	netLinkSigChan := make(chan int)

	paraWatcherSigChan := make(chan int)

	go linkParameterWatcher(paraWatcherSigChan)

	go netLinkDaemon(netOpInfo.RequestChann, netLinkSigChan, netOpInfo.ErrChan)
	watchChan := make(chan clientv3.WatchResponse)
	go func() {
		watch := utils.EtcdClient.Watch(context.Background(), key.NodeLinkListKeySelf)
		res := <-watch
		logrus.Infof("Etcd Instance Change Detected in Node %d", key.NodeIndex)
		watchChan <- res
	}()
	for {
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				netLinkSigChan <- sig
				paraWatcherSigChan <- sig
				return
			}
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Instance Info List Length:", len(res.Events))
				continue
			} else {
				logrus.Infof("Instance Change Detected in Node %d, list: %s", key.NodeIndex, string(res.Events[0].Kv.Value))
			}
			infoBytes := res.Events[0].Kv.Value
			updateIDList := []string{}
			err := json.Unmarshal(infoBytes, &updateIDList)
			if err != nil {
				logrus.Error("Parse Update Instance  String Info Error: ", err.Error())
			}
			addList, delList, err := parseLinkChange(updateIDList)
			if err != nil {
				logrus.Error("Parse Update Instance Info Error: ", err.Error())
			} else {
				logrus.Infof("Parse Update Instance Info Success: Addlist:%v,Dellist: %v", addList, delList)
			}
			err = DelLinks(delList, &netOpInfo)
			if err != nil {
				errMsg := fmt.Sprintf("Delete Containers %v Error: %s", delList, err.Error())
				logrus.Error(errMsg)
				errChan <- err

			}
			err = AddLinks(addList, &netOpInfo)
			if err != nil {
				errMsg := fmt.Sprintf("Add Containers %v Error: %s", delList, err.Error())
				logrus.Error(errMsg)
				errChan <- err
			}
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
