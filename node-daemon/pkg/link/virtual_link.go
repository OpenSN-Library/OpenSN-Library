package link

import (
	"NodeDaemon/data"
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const VirtualLinkType = "vlink"

const (
	VethDelayParameter     = "delay"
	VethLossParameter      = "loss"
	VethBandwidthParameter = "bandwidth"
)

var VirtualLinkParameterMap = map[string]model.ParameterInfo{
	model.ConnectParameter: model.ConnectParameterInfo,
	VethDelayParameter: {
		Name:           VethDelayParameter,
		MinVal:         0,
		MaxVal:         1e10,
		DefinitionFrac: 1e9,
		DefaultVal:     0,
	},
	VethLossParameter: {
		Name:           VethLossParameter,
		MinVal:         0,
		MaxVal:         10000,
		DefinitionFrac: 10000,
		DefaultVal:     0,
	},
	VethBandwidthParameter: {
		Name:           VethBandwidthParameter,
		MinVal:         0,
		MaxVal:         1e10,
		DefinitionFrac: 1,
		DefaultVal:     0,
	},
}

var TbfQdiscTemplate = netlink.Tbf{
	QdiscAttrs: netlink.QdiscAttrs{
		Handle: netlink.MakeHandle(2, 0),
		Parent: netlink.MakeHandle(1, 0),
	},
}
var NetemQdiscTemplate = netlink.NewNetem(
	netlink.QdiscAttrs{
		Handle: netlink.MakeHandle(1, 0),
		Parent: netlink.HANDLE_ROOT,
	},
	netlink.NetemQdiscAttrs{},
)

type VethLink struct {
	model.LinkBase
}

func CreateVethLinkObject(base model.LinkBase) *VethLink {
	return &VethLink{
		LinkBase: base,
	}
}

func (l *VethLink) IsCreated() bool {
	_, err := netlink.LinkByName(l.LinkID)

	return err == nil
}

func (l *VethLink) IsEnabled() bool {
	_, err := netlink.LinkByName(fmt.Sprintf("%s-%d", l.LinkID, 1))
	return err == nil
}

func (l *VethLink) Create() error {
	bridge := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   l.GetLinkID(),
			TxQLen: -1,
		},
	}

	err := netlink.LinkAdd(bridge)
	if err != nil {
		logrus.Errorf("Add Bridge Link %s Error: %s", bridge.Name, err.Error())
		return err
	}

	err = netlink.LinkSetUp(bridge)
	if err != nil {
		logrus.Errorf("Set Bridge Link %s Up Error: %s", bridge.Name, err.Error())
		return err
	}

	return err
}

func (l *VethLink) Destroy() error {
	logrus.Infof("Disabling Link %s, Type: Single Machine %s", l.LinkID, l.Type)
	bridge, err := netlink.LinkByName(l.GetLinkID())
	if err != nil {
		err := fmt.Errorf("get bridge device from name %s error: %s", l.LinkID, err.Error())
		return err
	}
	err = netlink.LinkDel(bridge)
	if err != nil {
		err := fmt.Errorf("delete bridge device %s error: %s", l.LinkID, err.Error())
		return err
	}
	return nil
}

func (l *VethLink) Connect() error {
	if !l.IsEnabled() {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.LinkID)
		return fmt.Errorf("%s is not enabled", l.LinkID)
	}

	setLink, err := netlink.LinkByName(l.GetLinkID())
	if err != nil {
		err := fmt.Errorf("get sub device from name %s error: %s", l.LinkID, err.Error())
		return err
	}
	err = netlink.LinkSetUp(setLink)
	if err != nil {
		err := fmt.Errorf("set sub device %s up error: %s", l.LinkID, err.Error())
		return err
	}

	logrus.Infof(
		"Connect Link %s Between %s and %s Success",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)
	return nil
}
func (l *VethLink) Disconnect() error {
	if !l.IsEnabled() {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.LinkID)
		return fmt.Errorf("%s is not enabled", l.LinkID)
	}

	setLink, err := netlink.LinkByName(l.LinkID)
	if err != nil {
		err := fmt.Errorf("get sub device from name %s error: %s", l.LinkID, err.Error())
		return err
	}
	err = netlink.LinkSetDown(setLink)
	if err != nil {
		err := fmt.Errorf("set sub device %s down error: %s", setLink.Attrs().Name, err.Error())
		return err
	}

	logrus.Infof(
		"Disconnect Link %s Between %s and %s Success",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)
	return nil
}

func (l *VethLink) enableSameMachine(brIndex int) error {

	for i, v := range l.EndInfos {
		instancePid := data.WatchInstancePid(v.InstanceID)
		veth := &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name:        fmt.Sprintf("%s-%d", l.GetLinkID(), i),
				MasterIndex: brIndex,
			},
			PeerName:      l.GetLinkID(),
			PeerNamespace: netlink.NsPid(instancePid),
		}

		err := netlink.LinkAdd(veth)
		if err != nil {
			logrus.Errorf("Add Veth Peer Link %v Error: %s", *l, err.Error())
		}
	}

	return nil
}

func (l *VethLink) enableCrossMachine(brIndex int) error {
	for i, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}
		targetNodeInfo, err := getNodeInfo(l.EndInfos[1-i].EndNodeIndex)
		if err != nil {
			return err
		}
		for i, v := range l.EndInfos {
			if v.EndNodeIndex != key.NodeIndex {
				vxlanDev := netlink.Vxlan{
					LinkAttrs: netlink.LinkAttrs{
						Name:        fmt.Sprintf("%s-%d", l.GetLinkID(), i),
						TxQLen:      -1,
						MasterIndex: brIndex,
					},
					VxlanId:  l.LinkIndex,
					SrcAddr:  key.SelfNode.L3AddrV4,
					Group:    targetNodeInfo.L3AddrV4,
					Port:     4789,
					Learning: true,
					L2miss:   true,
					L3miss:   true,
				}

				logrus.Infof("Create Vxlan %v", vxlanDev)
				err = netlink.LinkAdd(&vxlanDev)
				if err != nil {
					logrus.Errorf("Add Vxlan Link %v Error: %s", *l, err.Error())
				}
			} else {
				instancePid := data.WatchInstancePid(v.InstanceID)
				veth := &netlink.Veth{
					LinkAttrs: netlink.LinkAttrs{
						Name:        fmt.Sprintf("%s-%d", l.GetLinkID(), i),
						MasterIndex: brIndex,
					},
					PeerName:      l.GetLinkID(),
					PeerNamespace: netlink.NsPid(instancePid),
				}

				err = netlink.LinkAdd(veth)
				if err != nil {
					logrus.Errorf("Add Veth Peer Link %v Error: %s", *l, err.Error())
					return err
				}
			}
		}
	}
	return nil
}

func (l *VethLink) Enable() error {

	logrus.Infof("Enabling Link %s, Type: Single Machine %s", l.LinkID, l.Type)
	var err error
	bridge, err := netlink.LinkByName(l.LinkID)

	if err != nil {
		return fmt.Errorf("enable link error: get master bridge error: %s", err.Error())
	}

	if l.CrossMachine {
		err = l.enableCrossMachine(bridge.Attrs().Index)
	} else {
		err = l.enableSameMachine(bridge.Attrs().Index)
	}

	if err != nil {
		return err
	}

	return nil
}
func (l *VethLink) Disable() error {

	if !l.IsEnabled() {
		return nil
	}

	for i := range l.EndInfos {
		delLink, err := netlink.LinkByName(fmt.Sprintf("%s-%d", l.GetLinkID(), i))
		if err != nil {
			err := fmt.Errorf("get sub device from name %s error: %s", fmt.Sprintf("%s-%d", l.GetLinkID(), i), err.Error())
			return err
		}
		err = netlink.LinkDel(delLink)
		if err != nil {
			err := fmt.Errorf("delete sub device %s error: %s", delLink.Attrs().Name, err.Error())
			return err
		}
	}

	return nil
}

func (l *VethLink) SetParameters(oldPara, newPara map[string]int64) error {
	dirtyConnect := false
	dirtyTbf := false
	dirtyNetem := false

	for paraName, paraValue := range newPara {
		if paraValue == oldPara[paraName] {
			logrus.Debugf("Value of %s for %s is not changed, ignore.", paraName, l.LinkID)
			continue
		}
		if _, ok := VirtualLinkParameterMap[paraName]; !ok {
			logrus.Warnf("Unsupport Parameter %s for Link %s.", paraName, l.Type)
			continue
		}
		if paraValue != oldPara[paraName] {
			if paraName == model.ConnectParameter {
				dirtyConnect = true
			} else if paraName == VethBandwidthParameter {
				dirtyTbf = true
			} else {
				dirtyNetem = true
			}
		}
	}

	if dirtyConnect {
		if newPara[model.ConnectParameter] == 0 {
			err := l.Disconnect()
			if err != nil {
				logrus.Errorf("Disconnect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
			return err
		} else {
			err := l.Connect()
			if err != nil {
				logrus.Errorf("Connect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		}
	}

	for _, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}

		if dirtyNetem {
			for i, v := range l.EndInfos {
				if v.EndNodeIndex != key.NodeIndex {
					continue
				}
				dev, err := netlink.LinkByName(fmt.Sprintf("%s-%d", l.GetLinkID(), i))
				if err != nil {
					logrus.Errorf("Update netem qdisc error: get link by name %s error: %s", fmt.Sprintf("%s-%d", l.GetLinkID(), i), err.Error())
					return err
				}
				netemInfo := netlink.NewNetem(
					NetemQdiscTemplate.QdiscAttrs,
					netlink.NetemQdiscAttrs{
						Latency: uint32(newPara[VethDelayParameter]) + 1,
						Loss:    float32(newPara[VethLossParameter]) / 100,
					},
				)
				netemInfo.LinkIndex = dev.Attrs().Index
				err = netlink.QdiscReplace(netemInfo)
				if err != nil {
					logrus.Errorf("Update netem qdisc error: %s", err.Error())
				}
			}
		}

		if dirtyTbf {
			for i, v := range l.EndInfos {
				if v.EndNodeIndex != key.NodeIndex {
					continue
				}
				dev, err := netlink.LinkByName(fmt.Sprintf("%s-%d", l.GetLinkID(), i))
				if err != nil {
					logrus.Errorf("Update tbf qdisc error: get link by name %s error: %s", fmt.Sprintf("%s-%d", l.GetLinkID(), i), err.Error())
					return err
				}
				tbfInfo := TbfQdiscTemplate
				tbfInfo.LinkIndex = dev.Attrs().Index
				tbfInfo.Limit = uint32(newPara[VethBandwidthParameter])
				tbfInfo.Rate = uint64(newPara[VethBandwidthParameter])
				tbfInfo.Buffer = uint32(newPara[VethBandwidthParameter]) / 8
				err = netlink.QdiscReplace(&tbfInfo)
				if err != nil {
					logrus.Errorf("Update tbf qdisc error: %s", err.Error())
				}
			}
		}

	}

	return nil
}
