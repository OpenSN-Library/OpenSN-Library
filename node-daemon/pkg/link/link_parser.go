package link

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type DevInfoType struct {
	IfIndex int    `json:"if_index"`
	Name    string `json:"name"`
}

var linkIndexLock = new(sync.Mutex)

var LinkDeviceInfoMap = map[string][2]model.DeviceRequireInfo{
	VirtualLinkType: {
		{
			DevName: VirtualLinkType,
			NeedNum: 1,
			IsMutex: false,
		},
		{
			DevName: VirtualLinkType,
			NeedNum: 1,
			IsMutex: false,
		},
	},
	MultiplexPhysicalLinkType: {
		{
			DevName: MultiplexPhysicalLinkType,
			NeedNum: 1,
			IsMutex: false,
		},
		{
			DevName: MultiplexPhysicalLinkType,
			NeedNum: 1,
			IsMutex: false,
		},
	},
	FixPhysicalLinkType: {
		{
			DevName: FixPhysicalLinkType,
			NeedNum: 1,
			IsMutex: true,
		},
		{
			DevName: VirtualLinkType,
			NeedNum: 1,
			IsMutex: false,
		},
	},
}

func AllocLinkIndex() int {
	linkIndexLock.Lock()
	defer linkIndexLock.Unlock()
	index := 1
	concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {

		numStr := s.Get(key.NextLinkIndexKey)
		if numStr != "" {
			getIndex, err := strconv.Atoi(numStr)
			if err != nil {
				logrus.Errorf("Get next link index error: %s",err.Error())
			} else {
				index = getIndex
			}
			
		}

		s.Put(key.NextLinkIndexKey, strconv.Itoa(index+1))
		return nil
	})

	return index
}

func getNodeInfo(index int) (*model.Node, error) {
	nodeInfoKey := fmt.Sprintf("%s/%d", key.NodeIndexListKey, index)
	etcdNodeInfo, err := utils.EtcdClient.Get(
		context.Background(),
		nodeInfoKey,
	)

	if err != nil {
		err := fmt.Errorf("get node %d info from etcd error: %s", index, err.Error())
		return nil, err
	}

	if len(etcdNodeInfo.Kvs) <= 0 {
		return nil, fmt.Errorf("node %d not found", index)

	}

	v := new(model.Node)
	err = json.Unmarshal(etcdNodeInfo.Kvs[0].Value, v)
	if err != nil {
		err := fmt.Errorf(
			"unable to parse node info from etcd value %s, Error:%s",
			string(etcdNodeInfo.Kvs[0].Value),
			err.Error(),
		)
		return nil, err
	}
	return v, nil
}

func ParseLinkFromBytes(seq []byte) (model.Link, error) {
	var baseLink model.LinkBase
	var realLink model.Link
	err := json.Unmarshal(seq, &baseLink)
	if err != nil {
		logrus.Error("Unmarshal Json Data to Link Base Error, Redis Data May Crash: ", err.Error())
		return nil, err
	}
	realLink, err = ParseLinkFromBase(baseLink)
	return realLink, err
}

func ParseLinkFromBase(config model.LinkBase) (model.Link, error) {
	var realLink model.Link

	switch config.Type {
	case VirtualLinkType:
		vLink := CreateVethLinkObject(config)
		realLink = vLink

	case "":
		realLink = &VethLink{}
	default:
		err := fmt.Errorf("unsupported link type: %s", config.Type)
		logrus.Errorf("Parse Link Error: %s", err.Error())
		return nil, err
	}
	realLink.GetLinkBasePtr().CrossMachine = config.EndInfos[0].EndNodeIndex != config.EndInfos[1].EndNodeIndex
	return realLink, nil
}
