package biz

import (
	"NodeDaemon/config"
	"NodeDaemon/data"
	"NodeDaemon/model"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vishvananda/netlink"
)

type Parameter struct {
	BindInterfaceName string
	MasterNodeAddr    string
	NodeMode          string
}

func allocNodeIndex() error {
	status := utils.LockKeyWithTimeout(config.NextNodeIndexKey, 5*time.Second)
	if !status {
		return fmt.Errorf("unable to acquire lock of %s", config.NextNodeIndexKey)
	}
	getResp := utils.RedisClient.Get(context.Background(), config.NextNodeIndexKey)

	if getResp.Err() != nil && getResp.Err() != redis.Nil {
		return getResp.Err()
	} else {
		nodeIndex, err := strconv.Atoi(getResp.Val())
		if err != nil {
			return err
		}
		data.NodeIndex = nodeIndex
		setResp := utils.RedisClient.Set(context.Background(), config.NextNodeIndexKey, nodeIndex+1, 0)
		if setResp.Err() != nil {
			return setResp.Err()
		}
	}

	return nil
}

func UpdateNodeIndexList() error {

	var remoteIndexList []int = []int{}

	status := utils.LockKeyWithTimeout(config.NodeIndexListKey, 6*time.Second)
	if !status {
		return fmt.Errorf("unable to access lock of key %s", config.NodeIndexListKey)
	}
	getResp, err := utils.EtcdClient.Get(context.Background(), config.NodeIndexListKey)
	if err != nil {
		return err
	}

	if len(getResp.Kvs) >= 1 {
		err = json.Unmarshal(getResp.Kvs[0].Value, &remoteIndexList)

		if err != nil {
			return err
		}
	}

	remoteIndexList = append(remoteIndexList, data.NodeIndex)

	updateBytes, err := json.Marshal(remoteIndexList)

	if err != nil {
		return err
	}

	_, err = utils.EtcdClient.Put(context.Background(), config.NodeIndexListKey, string(updateBytes))

	return err
}

func byteSeqEncode(b []byte) (ret uint64) {
	for i := 0; i < len(b); i++ {
		ret <<= 2
		if i < len(b) {
			ret |= (uint64(b[i]) & 0xff)
		}
	}
	return
}

func getInterfaceInfo(ifName string, target *model.Node) error {
	var link netlink.Link
	var err error
	if ifName == "" {
		linkList, err := netlink.RouteList(nil, 4)
		if err != nil {
			return err
		}
		if len(linkList) <= 0 {
			return errors.New("unable to find default route interface")
		}
		ifIndex := linkList[0].LinkIndex
		link, err = netlink.LinkByIndex(ifIndex)
		if err != nil {
			return err
		}
	} else {
		link, err = netlink.LinkByName(ifName)
		if err != nil {
			return err
		}
	}
	target.L2Addr = byteSeqEncode(link.Attrs().HardwareAddr)
	linkV4Addrs, err := netlink.AddrList(link, 4)
	if err != nil {
		return err
	}
	if len(linkV4Addrs) > 0 {
		target.L3AddrV4 = uint32(byteSeqEncode(linkV4Addrs[0].IP))
	}

	linkV6Addrs, err := netlink.AddrList(link, 6)
	if err != nil {
		return err
	}
	if len(linkV4Addrs) > 0 {
		target.L3AddrV6 = byteSeqEncode(linkV6Addrs[0].IP)
	}
	return nil
}

func NodeInit() error {

	if config.StartMode == config.MasterNode {
		data.NodeIndex = 0
		err := config.InitConfigMasterMode()
		if err != nil {
			return err
		}
	} else {
		err := config.InitConfigServantMode(config.MasterAddress)
		if err != nil {
			return err
		}
		err = allocNodeIndex()
		if err != nil {
			return fmt.Errorf("alloc node index error: %s", err.Error())
		}
	}

	selfInfo := &model.Node{
		NodeID:        uint32(data.NodeIndex),
		FreeInstance:  model.MAX_INSTANCE_NODE,
		IsMasterNode:  data.NodeIndex == 0,
		NsInstanceMap: map[string]string{},
		NsLinkMap:     map[string]string{},
	}

	if data.NodeIndex == 0 {
		selfInfo.FreeInstance -= model.MASTER_NODE_MAKEUP
	}

	err := getInterfaceInfo(config.InterfaceName, selfInfo)

	if err != nil {
		return fmt.Errorf("get interface info error: %s", err.Error())
	}

	selfInfoBytes, err := json.Marshal(selfInfo)

	if err != nil {
		return fmt.Errorf("marshal node info error: %s", err.Error())
	}

	setResp := utils.RedisClient.HSet(
		context.Background(),
		config.NodesKey,
		strconv.Itoa(data.NodeIndex),
		string(selfInfoBytes),
	)

	if setResp.Err() != nil {
		return fmt.Errorf("update node info error: %s", setResp.Err().Error())
	}

	err = UpdateNodeIndexList()
	if err != nil {
		return fmt.Errorf("update node list error: %s", err.Error())
	}

	return nil
}
