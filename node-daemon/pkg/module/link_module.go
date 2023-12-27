package module

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/pkg/link"
	"NodeDaemon/share/data"
	"NodeDaemon/share/dir"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

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
		errMsg := fmt.Sprintf("Check Node Link List Initialized %s Error: %s", key.NodeLinkListKeySelf, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	if len(getResp.Kvs) <= 0 {
		_, err := utils.EtcdClient.Put(
			context.Background(),
			key.NodeLinkListKeySelf,
			"[]",
		)
		if err != nil {
			errMsg := fmt.Sprintf("Init Node Link List %s Error: %s", key.NodeLinkListKeySelf, err.Error())
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
			logrus.Error("Get Link Infos Error: ", err.Error())
			return
		}

		for i, v := range redisResponse.Val() {
			if v == nil {
				logrus.Error("Redis Result Empty, Redis Data May Crash, LinkID:", addList[i])
				continue
			} else {
				newLink, err := link.ParseLinkFromBytes([]byte(v.(string)))
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
				logrus.Errorf("Get net namespace from pid %d error: %s", req.GetLinkNamespacePid(), err.Error())
				errChan <- err
				continue
			}
			err = netns.Set(linkNsFd)
			if err != nil {
				logrus.Errorf("Set net namespace error: %s", err.Error())
				errChan <- err
				continue
			}
			link, err := netlink.LinkByName(req.GetLinkName())
			if err != nil {
				logrus.Errorf("Get link from index %d error: %s", req.GetLinkIndex(), err.Error())
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
				addr := strings.Split(realReq.V4Addr, "/")
				if len(addr) < 2 {
					opErr = fmt.Errorf("invalid ipv4 addr %s", realReq.V4Addr)
					break
				}
				ip := net.ParseIP(addr[0])
				prefixLen, err := strconv.Atoi(addr[1])
				if err != nil {
					opErr = fmt.Errorf("invalid ipv4 addr prefix length %s", err.Error())
					break
				}

				netlinkAddr := netlink.Addr{
					IPNet: &net.IPNet{
						IP:   ip,
						Mask: utils.CreateV4InetMask(prefixLen),
					},
				}
				opErr = netlink.AddrAdd(link, &netlinkAddr)
			case netreq.SetV6Addr:
				realReq := req.(*netreq.SetV6AddrReq)
				addr := strings.Split(realReq.V6Addr, "/")
				if len(addr) < 2 {
					opErr = fmt.Errorf("invalid ipv4 addr %s", realReq.V6Addr)
					break
				}
				ip := net.ParseIP(addr[0])
				prefixLen, err := strconv.Atoi(addr[1])
				if err != nil {
					opErr = fmt.Errorf("invalid ipv4 addr %s", realReq.V6Addr)
					break
				}

				netlinkAddr := netlink.Addr{
					IPNet: &net.IPNet{
						IP:   ip,
						Mask: utils.CreateV6InetMask(prefixLen),
					},
				}
				opErr = netlink.AddrAdd(link, &netlinkAddr)
			case netreq.SetNetNs:
				realReq := req.(*netreq.SetNetNsReq)
				opErr = netlink.LinkSetNsPid(link, realReq.TargetNamespacePid)
			case netreq.SetQdisc:
				realReq := req.(*netreq.SetQdiscReq)
				opErr = netlink.QdiscReplace(realReq.QdiscInfo)
			case netreq.DeleteLink:
				opErr = netlink.LinkDel(link)
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
			if opErr != nil {
				logrus.Errorf("Netlink operation Error, Type %d, error: %s", req.GetRequestType(), opErr.Error())
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

func updateTopoInfoFile(addList []string, delList []model.Link) error {
	dirtyMap := make(map[string]bool)
	for _, v := range addList {
		linkConfig := data.LinkMap[v].GetLinkConfig()
		for i, instanceID := range linkConfig.InitInstanceID {
			targetIndex := 1 - i%2
			targetInstanceID := linkConfig.InitInstanceID[targetIndex]
			if targetInstanceID == "" {
				continue
			} else {
				dirtyMap[instanceID] = true
			}
			if topoInfo, ok := data.TopoInfoMap[instanceID]; ok {

				topoInfo.LinkInfos[targetInstanceID] = &model.LinkInfo{
					V4Addr: linkConfig.IPInfos[targetIndex].V4Addr,
					V6Addr: linkConfig.IPInfos[targetIndex].V6Addr,
				}
				topoInfo.EndInfos[targetInstanceID] = &model.EndInfo{
					InstanceID: targetInstanceID,
					Type:       data.InstanceMap[targetInstanceID].Config.Type,
				}
			} else {

				data.TopoInfoMap[instanceID] = &model.TopoInfo{
					InstanceID: instanceID,
					LinkInfos: map[string]*model.LinkInfo{
						targetInstanceID: {
							V4Addr: linkConfig.IPInfos[targetIndex].V4Addr,
							V6Addr: linkConfig.IPInfos[targetIndex].V6Addr,
						},
					},
					EndInfos: map[string]*model.EndInfo{
						targetInstanceID: {
							InstanceID: targetInstanceID,
							Type:       data.InstanceMap[targetInstanceID].Config.Type,
						},
					},
				}

			}
		}
	}

	for _, v := range delList {
		linkConfig := v.GetLinkConfig()
		for i, instanceID := range linkConfig.InitInstanceID {
			targetIndex := 1 - i%2
			targetInstanceID := linkConfig.InitInstanceID[targetIndex]
			if targetInstanceID == "" {
				continue
			} else {
				dirtyMap[instanceID] = true
			}

			delete(data.TopoInfoMap[instanceID].EndInfos, targetInstanceID)
			delete(data.TopoInfoMap[instanceID].LinkInfos, targetInstanceID)

			if len(data.TopoInfoMap[instanceID].LinkInfos) == 0 {
				delete(data.TopoInfoMap, instanceID)
			}
		}
	}

	for instanceID := range dirtyMap {
		jsonPath := path.Join(dir.TopoInfoDir, fmt.Sprintf("%s.json", instanceID))
		if topoInfo, ok := data.TopoInfoMap[instanceID]; ok {
			fileContent, _ := json.Marshal(topoInfo)
			err := utils.WriteToFile(jsonPath, fileContent)
			if err != nil {
				logrus.Errorf("Write Topo Infomation of %s to path %s Error: %s", instanceID, jsonPath, err.Error())
				return err
			}
		} else {
			err := utils.DeleteFile(jsonPath)
			if err != nil {
				logrus.Errorf("Delete Topo Infomation of %s from path %s Error: %s", instanceID, jsonPath, err.Error())
				return err
			}
		}
	}

	return nil
}

func linkParameterWatcher(sigChan chan int) {
	for {
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				return
			}
		}
	}
}

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	Base
}

func AddLinks(addList []string, operator *model.NetlinkOperatorInfo) error {
	for _, v := range addList {
		linkInfo := data.LinkMap[v]
		utils.Spin(func() bool {
			res := true
			for _, v := range linkInfo.GetEndInfos() {
				if v.InstanceID == "" {
					continue
				}
				instanceInfo, ok := data.InstanceMap[v.InstanceID]
				if !ok || instanceInfo.Pid == 0 {
					res = false
					break
				}
			}
			return res
		}, 100*time.Millisecond)

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
		logrus.Infof("Deleting Link %s: %v", v.GetLinkID(), v)
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
	for {
		go func() {
			watch := utils.EtcdClient.Watch(context.Background(), key.NodeLinkListKeySelf)
			res := <-watch
			logrus.Infof("Etcd Link Change Detected in Node %d", key.NodeIndex)
			watchChan <- res
		}()
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				netLinkSigChan <- sig
				paraWatcherSigChan <- sig
				return
			}
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Link Info List Length:", len(res.Events))
				continue
			} else {
				logrus.Infof("Link Change Detected in Node %d, list: %s", key.NodeIndex, string(res.Events[0].Kv.Value))
			}
			infoBytes := res.Events[0].Kv.Value
			updateIDList := []string{}
			err := json.Unmarshal(infoBytes, &updateIDList)
			if err != nil {
				logrus.Error("Parse Update Link  String Info Error: ", err.Error())
			}
			addList, delList, err := parseLinkChange(updateIDList)
			if err != nil {
				logrus.Error("Parse Update Link Info Error: ", err.Error())
			} else {
				logrus.Infof("Parse Update Link Info Success: Addlist:%v,Dellist: %v", addList, delList)
			}
			go func() {
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

				err = updateTopoInfoFile(addList, delList)

				if err != nil {
					errMsg := fmt.Sprintf("Update Container Topo Infomation Error: %s", err.Error())
					logrus.Error(errMsg)
					errChan <- err
				}

			}()
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
