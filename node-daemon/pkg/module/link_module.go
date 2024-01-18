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
	"path"

	"sync"

	"github.com/sirupsen/logrus"
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

func updateTopoInfoFile(addList []string, delList []model.Link) error {
	dirtyMap := make(map[string]bool)
	for _, v := range addList {
		linkConfig := data.LinkMap[v].GetLinkConfig()
		for i, endInfo := range linkConfig.InitEndInfos {
			targetIndex := 1 - i%2
			targetInstanceID := linkConfig.InitEndInfos[targetIndex].InstanceID
			if targetInstanceID == "" {
				continue
			} else {
				dirtyMap[endInfo.InstanceID] = true
			}
			if topoInfo, ok := data.TopoInfoMap[endInfo.InstanceID]; ok {

				topoInfo.LinkInfos[targetInstanceID] = &model.LinkInfo{
					V4Addr: linkConfig.IPInfos[targetIndex].V4Addr,
					V6Addr: linkConfig.IPInfos[targetIndex].V6Addr,
				}
				topoInfo.EndInfos[targetInstanceID] = &model.EndInfo{
					InstanceID: targetInstanceID,
					Type:       linkConfig.InitEndInfos[targetIndex].InstanceType,
				}
			} else {

				data.TopoInfoMap[endInfo.InstanceID] = &model.TopoInfo{
					InstanceID: endInfo.InstanceID,
					LinkInfos: map[string]*model.LinkInfo{
						targetInstanceID: {
							V4Addr: linkConfig.IPInfos[targetIndex].V4Addr,
							V6Addr: linkConfig.IPInfos[targetIndex].V6Addr,
						},
					},
					EndInfos: map[string]*model.EndInfo{
						targetInstanceID: {
							InstanceID: targetInstanceID,
							Type:       linkConfig.InitEndInfos[targetIndex].InstanceType,
						},
					},
				}

			}
		}
	}

	for _, v := range delList {
		linkConfig := v.GetLinkConfig()
		for i, endInfo := range linkConfig.InitEndInfos {
			targetIndex := 1 - i%2
			targetInstanceID := linkConfig.InitEndInfos[targetIndex].InstanceID
			if targetInstanceID == "" || data.InstanceMap[targetInstanceID] == nil {
				continue
			} else {
				dirtyMap[endInfo.InstanceID] = true
			}

			delete(data.TopoInfoMap[endInfo.InstanceID].EndInfos, targetInstanceID)
			delete(data.TopoInfoMap[endInfo.InstanceID].LinkInfos, targetInstanceID)

			if len(data.TopoInfoMap[endInfo.InstanceID].LinkInfos) == 0 {
				delete(data.TopoInfoMap, endInfo.InstanceID)
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

func linkParameterWatcher(sigChan chan int, operator *model.NetlinkOperatorInfo) {
	ctx, cancel := context.WithCancel(context.Background())

	watchChan := utils.EtcdClient.Watch(ctx, key.NodeLinkParameterKeySelf)
	for {
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				cancel()
				return
			}
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Link Parameter Info List Length:", len(res.Events))
				continue
			}
			infoBytes := res.Events[0].Kv.Value
			newLinkParameter := make(map[string]map[string]int64)
			err := json.Unmarshal(infoBytes, &newLinkParameter)
			if err != nil {
				logrus.Error("Parse Update Link Parameter String Info Error: ", err.Error())
			}
			for linkID, parameter := range newLinkParameter {
				if link2Update, ok := data.LinkMap[linkID]; ok {
					if !link2Update.IsEnabled() {
						continue
					}
					reqs, err := link2Update.SetParameters(parameter)
					if err != nil {
						logrus.Errorf("Generate Update Parameter Requests of Link %s Error: %s", linkID, err.Error())
						continue
					}
					operator.RequestChann <- reqs
				}
			}
		}
	}
}

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	Base
}

func AddLinks(addList []string, operator *model.NetlinkOperatorInfo) error {

	err := utils.ForEachUtilAllComplete[string](
		func(v string) (bool, error) {

			linkInfo := data.LinkMap[v]
			for _, v := range linkInfo.GetEndInfos() {
				if v.InstanceID == "" {
					return true, nil
				}
				instanceInfo, ok := data.InstanceMap[v.InstanceID]
				if ok && instanceInfo.Pid == 0 {
					return false, nil
				}
			}

			requests, err := linkInfo.Enable()
			if err != nil {
				logrus.Errorf("Generate Enable Link %s Requests Error: %s", linkInfo.GetLinkID(), err.Error())
				return true, err
			}
			connectReqs, err := linkInfo.Connect()
			if err != nil {
				logrus.Errorf("Connect Link %s Error: %s", linkInfo.GetLinkID(), err.Error())
				return true, err
			}
			requests = append(requests, connectReqs...)
			operator.RequestChann <- requests
			logrus.Infof("Enable and Connect Link %s Between %s and %s, Type %s",
				linkInfo.GetLinkID(),
				linkInfo.GetLinkConfig().InitEndInfos[0].InstanceID,
				linkInfo.GetLinkConfig().InitEndInfos[1].InstanceID,
				linkInfo.GetLinkType(),
			)
			return true, nil
		}, addList,
	)
	if err != nil {
		logrus.Errorf("Add Links Error: %s", err.Error())
		return err
	}
	return nil
}

func DelLinks(delList []model.Link, operator *model.NetlinkOperatorInfo) error {

	utils.ForEachWithThreadPool[model.Link](
		func(v model.Link) {
			var request []netreq.NetLinkRequest
			logrus.Infof("Deleting Link %s: %v", v.GetLinkID(), v)
			if v.IsConnected() {
				disconnReqs, err := v.Disconnect()
				if err != nil {
					logrus.Errorf("Disconnect Link %s Error: %s", v.GetLinkID(), err.Error())
				}
				request = append(request, disconnReqs...)
			}
			disableReqs, err := v.Disable()
			if err != nil {
				logrus.Errorf("Disable Link %s Error: %s", v.GetLinkID(), err.Error())
			}
			request = append(request, disableReqs...)
			operator.RequestChann <- request
			delete(data.LinkMap, v.GetLinkID())
		}, delList, 32,
	)

	return nil
}

func linkDaemonFunc(sigChan chan int, errChan chan error) {
	InitLinkData()
	operatorNum := 32
	netOpInfo := model.NetlinkOperatorInfo{
		RequestChann: make(chan []netreq.NetLinkRequest, operatorNum),
	}
	netLinkSigChan := make(chan int)
	paraWatcherSigChan := make(chan int)

	go linkParameterWatcher(paraWatcherSigChan, &netOpInfo)
	for i := 0; i < operatorNum; i++ {
		go NetLinkOperator(netOpInfo.RequestChann, netLinkSigChan, i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	watchChan := utils.EtcdClient.Watch(ctx, key.NodeLinkListKeySelf)
	for {

		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				cancel()
				netLinkSigChan <- sig
				paraWatcherSigChan <- sig
				return
			}
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Link Info List Length:", len(res.Events))
				continue
			}
			infoBytes := res.Events[0].Kv.Value
			updateIDList := []string{}
			err := json.Unmarshal(infoBytes, &updateIDList)
			if err != nil {
				logrus.Error("Parse Update Link String Info Error: ", err.Error())
			}
			addList, delList, err := parseLinkChange(updateIDList)
			if err != nil {
				logrus.Error("Parse Update Link Info Error: ", err.Error())
			} else {
				logrus.Infof("Parse Update Link Info Success: Addlist:%v,Dellist: %v", addList, delList)
			}
			err = updateTopoInfoFile(addList, delList)

			if err != nil {
				errMsg := fmt.Sprintf("Update Container Topo Infomation Error: %s", err.Error())
				logrus.Error(errMsg)
				errChan <- err
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
