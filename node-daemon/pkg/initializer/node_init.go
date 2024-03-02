package initializer

import (
	"NodeDaemon/config"
	"NodeDaemon/model"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"errors"
	"fmt"
	"strconv"

	"github.com/vishvananda/netlink"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Parameter struct {
	BindInterfaceName string
	MasterNodeAddr    string
	NodeMode          string
}

func allocNodeIndex() error {

	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		indexStr := s.Get(key.NextNodeIndexKey)

		if len(indexStr) <= 0 {
			key.NodeIndex = 0
		} else {
			nodeIndex, err := strconv.Atoi(indexStr)
			if err != nil {
				return err
			}
			key.NodeIndex = nodeIndex

		}
		s.Put(key.NextNodeIndexKey, strconv.Itoa(key.NodeIndex+1))
		return nil
	})
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
	if len(linkV6Addrs) > 0 {
		target.L3AddrV6 = linkV6Addrs[0].IP
	}
	return nil
}

func NodeInit() error {
	err := InitWorkdir()
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
	}

	err = utils.InitEtcdClient(
		config.GlobalConfig.Dependency.EtcdAddr,
		config.GlobalConfig.Dependency.EtcdPort,
	)
	if err != nil {
		return err
	}
	err = utils.InitDockerClient(config.GlobalConfig.Dependency.DockerHostPath)
	if err != nil {
		return err
	}

	if config.GlobalConfig.App.EnableMonitor {
		err := utils.InitInfluxDB(
			config.GlobalConfig.Dependency.InfluxdbAddr,
			config.GlobalConfig.Dependency.InfluxdbToken,
			config.GlobalConfig.Dependency.InfluxdbOrg,
			config.GlobalConfig.Dependency.InfluxdbBucket,
			config.GlobalConfig.Dependency.InfluxdbPort,
		)
		if err != nil {
			return err
		}
	}
	err = allocNodeIndex()
	if err != nil {
		return fmt.Errorf("alloc node index error: %s", err.Error())
	}

	key.InitKeys()
	selfInfo := &model.Node{
		NodeIndex:          key.NodeIndex,
		FreeInstance:       config.GlobalConfig.App.InstanceCapacity,
		IsMasterNode:       key.NodeIndex == 0,
		NsInstanceMap:      map[string]string{},
		NsLinkMap:          map[string]string{},
		NodeLinkDeviceInfo: map[string]int{},
	}

	for k, v := range config.GlobalConfig.Device {
		selfInfo.NodeLinkDeviceInfo[k] = len(v)
	}

	err = getInterfaceInfo(config.GlobalConfig.App.InterfaceName, selfInfo)

	if err != nil {
		return fmt.Errorf("get interface info error: %s", err.Error())
	}
	key.SelfNode = selfInfo
	err = synchronizer.AddNode(selfInfo)

	if err != nil {
		return fmt.Errorf("Add node %d error: %s", selfInfo.NodeIndex, err.Error())
	}
	return nil
}
