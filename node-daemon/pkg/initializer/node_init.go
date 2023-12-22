package initializer

import (
	"NodeDaemon/config"
	"NodeDaemon/model"
	"NodeDaemon/share/key"
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
	status := utils.LockKeyWithTimeout(key.NextNodeIndexKey, 5*time.Second)
	if !status {
		return fmt.Errorf("unable to acquire lock of %s", key.NextNodeIndexKey)
	}
	getResp := utils.RedisClient.Get(context.Background(), key.NextNodeIndexKey)

	if getResp.Err() != nil && getResp.Err() != redis.Nil {
		return getResp.Err()
	} else {
		nodeIndex, err := strconv.Atoi(getResp.Val())
		if err != nil {
			return err
		}
		key.NodeIndex = nodeIndex
		setResp := utils.RedisClient.Set(context.Background(), key.NextNodeIndexKey, nodeIndex+1, 0)
		if setResp.Err() != nil {
			return setResp.Err()
		}
	}

	return nil
}

func UpdateNodeIndexList() error {

	var remoteIndexList []int = []int{}

	status := utils.LockKeyWithTimeout(key.NodeIndexListKey, 6*time.Second)
	if !status {
		return fmt.Errorf("unable to access lock of key %s", key.NodeIndexListKey)
	}
	getResp, err := utils.EtcdClient.Get(context.Background(), key.NodeIndexListKey)
	if err != nil {
		return err
	}

	if len(getResp.Kvs) >= 1 {
		err = json.Unmarshal(getResp.Kvs[0].Value, &remoteIndexList)

		if err != nil {
			return err
		}
	}

	remoteIndexList = append(remoteIndexList, key.NodeIndex)

	updateBytes, err := json.Marshal(remoteIndexList)

	if err != nil {
		return err
	}

	_, err = utils.EtcdClient.Put(context.Background(), key.NodeIndexListKey, string(updateBytes))

	return err
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
	target.L2Addr = link.Attrs().HardwareAddr
	linkV4Addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	if len(linkV4Addrs) > 0 {
		target.L3AddrV4 = linkV4Addrs[0].IP
	}

	linkV6Addrs, err := netlink.AddrList(link, netlink.FAMILY_V6)
	if err != nil {
		return err
	}
	if len(linkV4Addrs) > 0 {
		target.L3AddrV6 = linkV6Addrs[0].IP
	}
	return nil
}

func NodeInit() error {
	err := utils.InitEtcdClient(
		config.GlobalConfig.Dependency.EtcdAddr,
		config.GlobalConfig.Dependency.EtcdPort,
	)
	if err != nil {
		return err
	}
	utils.InitRedisClient(
		config.GlobalConfig.Dependency.RedisAddr,
		config.GlobalConfig.Dependency.RedisPassword,
		config.GlobalConfig.Dependency.RedisPort,
		config.GlobalConfig.Dependency.RedisDBIndex,
	)
	err = utils.InitDockerClient(config.GlobalConfig.Dependency.DockerHostPath)
	if err != nil {
		return err
	}
	if !config.GlobalConfig.App.IsServant {
		key.NodeIndex = 0
		err := config.InitConfigMasterMode()
		if err != nil {
			return err
		}
	} else {
		err := config.InitConfigServantMode(config.GlobalConfig.App.MasterAddress)
		if err != nil {
			return err
		}
		err = allocNodeIndex()
		if err != nil {
			return fmt.Errorf("alloc node index error: %s", err.Error())
		}
	}
	key.InitKeys()
	selfInfo := &model.Node{
		NodeID:        uint32(key.NodeIndex),
		FreeInstance:  model.MAX_INSTANCE_NODE,
		IsMasterNode:  key.NodeIndex == 0,
		NsInstanceMap: map[string]string{},
		NsLinkMap:     map[string]string{},
	}

	for k, v := range config.GlobalConfig.Device {
		selfInfo.NodeLinkDeviceInfo[k] = len(v)
	}

	if key.NodeIndex == 0 {
		selfInfo.FreeInstance -= model.MASTER_NODE_MAKEUP
	}

	err = getInterfaceInfo(config.GlobalConfig.App.InterfaceName, selfInfo)

	if err != nil {
		return fmt.Errorf("get interface info error: %s", err.Error())
	}

	selfInfoBytes, err := json.Marshal(selfInfo)

	if err != nil {
		return fmt.Errorf("marshal node info error: %s", err.Error())
	}

	setResp := utils.RedisClient.HSet(
		context.Background(),
		key.NodesKey,
		strconv.Itoa(key.NodeIndex),
		string(selfInfoBytes),
	)

	if setResp.Err() != nil {
		return fmt.Errorf("update node info error: %s", setResp.Err().Error())
	}

	err = UpdateNodeIndexList()
	if err != nil {
		return fmt.Errorf("update node list error: %s", err.Error())
	}
	return InitData()
}
