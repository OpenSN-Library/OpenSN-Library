package link

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const VirtualLinkType = "VirtualLink"

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

type DevInfoType struct {
	IfIndex int    `json:"if_index"`
	Name    string `json:"name"`
}

type VethLink struct {
	model.LinkBase

	DevInfos           [2]DevInfoType `json:"dev_info"`
	TbfQdiscTemplate   netlink.Tbf
	NetemQdiscTemplate netlink.Netem
}

func CreateVethLinkObject(initConfig model.LinkConfig) *VethLink {

	return &VethLink{
		LinkBase: model.LinkBase{
			Enabled:           false,
			CrossMachine:      false,
			SupportParameters: VirtualLinkParameterMap,
			Parameter:         initConfig.InitParameter,
			Config:            initConfig,
			EndInfos:          initConfig.InitEndInfos,
		},
		DevInfos: [2]DevInfoType{
			{
				Name:    fmt.Sprintf("%s_0", initConfig.LinkID),
				IfIndex: -1,
			},
			{
				Name:    fmt.Sprintf("%s_1", initConfig.LinkID),
				IfIndex: -1,
			},
		},
		NetemQdiscTemplate: *netlink.NewNetem(
			netlink.QdiscAttrs{
				Handle: netlink.MakeHandle(1, 0),
				Parent: netlink.HANDLE_ROOT,
			},
			netlink.NetemQdiscAttrs{},
		),
		TbfQdiscTemplate: netlink.Tbf{
			QdiscAttrs: netlink.QdiscAttrs{
				Handle: netlink.MakeHandle(2, 0),
				Parent: netlink.MakeHandle(1, 0),
			},
		},
	}
}

func (l *VethLink) Connect(operatorInfo *model.NetlinkOperatorInfo) error {
	if !l.Enabled {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}
	for i, v := range l.EndInfos {
		if v.InstanceID == "" {
			logrus.Infof("Skip Link %s, because it's float link", l.Config.LinkID)
		}
		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if !ok {
			logrus.Errorf("Connect %s and %s Error: Instance %s is not in Node %d", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, v.InstanceID, key.NodeIndex)
			return fmt.Errorf("%s is not in node %d", v.InstanceID, key.NodeIndex)
		}
		setStateReq := netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, true)
		operatorInfo.RequestChann <- &setStateReq
		err := <-operatorInfo.ErrChan
		if err != nil {
			logrus.Errorf("Connect %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
			return err
		}
	}
	logrus.Infof(
		"Connect Link %s Between %s and %s Suucess",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)

	return nil
}
func (l *VethLink) Disconnect(operatorInfo *model.NetlinkOperatorInfo) error {
	if !l.Enabled {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}
	for i, v := range l.EndInfos {

		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if !ok {
			logrus.Errorf("Disconnect %s and %s Error: Instance %s is not in Node %d", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, v.InstanceID, key.NodeIndex)
			return fmt.Errorf("%s is not in node %d", v.InstanceID, key.NodeIndex)
		}
		setStateReq := netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, false)
		operatorInfo.RequestChann <- &setStateReq
		err := <-operatorInfo.ErrChan
		if err != nil {
			if err != nil {
				logrus.Errorf("Disconnect %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		}
	}
	logrus.Infof(
		"Disconnect Link %s Between %s and %s Suucess",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)
	return nil
}

func (l *VethLink) enableSameMachine() error {
	logrus.Infof("Enabling Link %s ,Type: Single Machine %s", l.Config.LinkID, l.Config.Type)
	if l.Enabled {
		logrus.Errorf("Enable %s and %s Error: %s has already been enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return fmt.Errorf("%s is enabled", l.Config.LinkID)
	}

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: l.DevInfos[0].Name,
		},
		PeerName: l.DevInfos[1].Name,
	}

	err := netlink.LinkAdd(veth)

	if err != nil {
		logrus.Errorf("Add Veth Peer Link Error: %s", err.Error())
		return err
	}

	l.DevInfos[0].IfIndex = veth.Index
	peerIndex, err := netlink.VethPeerIndex(veth)
	if err != nil {
		logrus.Errorf("Get Peer Link Index of %d Error: %s", veth.Attrs().Index, err.Error())
		return err
	}

	l.DevInfos[1].IfIndex = peerIndex
	if err != nil {
		logrus.Errorf("Get Init Net Namespace Error: %s", err.Error())
	}

	return nil
}

func (l *VethLink) enableCrossMachine() error {
	for i, v := range l.EndInfos {
		_, ok := data.InstanceMap[v.InstanceID]
		if ok {
			vxlanDev := netlink.Vxlan{
				LinkAttrs: netlink.LinkAttrs{
					Name:   l.DevInfos[i].Name,
					TxQLen: -1,
				},
				VxlanId:  0,
				SrcAddr:  data.NodeMap[key.NodeIndex].L3AddrV4,
				Group:    data.NodeMap[l.EndInfos[i%1].EndNodeIndex].L3AddrV4,
				Port:     0,
				Learning: true,
				L2miss:   true,
				L3miss:   true,
			}
			err := netlink.LinkAdd(&vxlanDev)
			if err != nil {
				return err
			}
			l.DevInfos[i].IfIndex = vxlanDev.Index
		}
	}

	return nil
}

func (l *VethLink) Enable(operatorInfo *model.NetlinkOperatorInfo) error {
	var err error
	if l.CrossMachine {
		err = l.enableCrossMachine()
	} else {
		err = l.enableSameMachine()
	}

	if err != nil {
		return err
	}

	for i, v := range l.EndInfos {
		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if v.InstanceID == "" {
			continue
		}
		if !ok {
			logrus.Errorf("Enable Link Between %s and %s Error: Instance %s is not in Node %d", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, v.InstanceID, key.NodeIndex)
			return fmt.Errorf("%s is not in node %d", v.InstanceID, key.NodeIndex)
		}

		setNsreq := netreq.CreateSetNetNsReq(l.DevInfos[i].IfIndex, os.Getpid(), instanceInfo.Pid, l.DevInfos[i].Name)
		operatorInfo.RequestChann <- &setNsreq
		err := <-operatorInfo.ErrChan
		if err != nil {
			logrus.Errorf("Set Link Netns %d Error: %s", l.DevInfos[i].IfIndex, err.Error())
			return err
		}
		setStateReq := netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, true)
		operatorInfo.RequestChann <- &setStateReq
		err = <-operatorInfo.ErrChan
		if err != nil {
			logrus.Errorf("Set Link %d Up Error: %s", l.DevInfos[i].IfIndex, err.Error())
			return err
		}
		if len(l.Config.IPInfos[i].V4Addr) > 0 {
			setV4AddrReq := netreq.CreateSetV4AddrReq(
				l.DevInfos[i].IfIndex,
				instanceInfo.Pid,
				l.DevInfos[i].Name,
				l.Config.IPInfos[i].V4Addr,
			)
			operatorInfo.RequestChann <- &setV4AddrReq
			err = <-operatorInfo.ErrChan
			if err != nil {
				logrus.Errorf("Set Link %d V4 Addr %s Error: %s", l.DevInfos[i].IfIndex, l.Config.IPInfos[i].V4Addr, err.Error())
				return err
			}
		}
		if len(l.Config.IPInfos[i].V6Addr) > 0 {
			setV6AddrReq := netreq.CreateSetV6AddrReq(
				l.DevInfos[i].IfIndex,
				instanceInfo.Pid,
				l.DevInfos[i].Name,
				l.Config.IPInfos[i].V6Addr,
			)
			operatorInfo.RequestChann <- &setV6AddrReq
			err = <-operatorInfo.ErrChan
			if err != nil {
				logrus.Errorf("Set Link %d V6 Addr %s Error: %s", l.DevInfos[i].IfIndex, l.Config.IPInfos[i].V6Addr, err.Error())
				return err
			}
		}
		err = l.SetParameters(l.Config.InitParameter, operatorInfo)
		if err != nil {
			return err
		}
	}
	l.Enabled = true
	return nil
}
func (l *VethLink) Disable(operatorInfo *model.NetlinkOperatorInfo) error {
	if !l.Enabled {
		logrus.Errorf("Disable %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}
	instanceInfo, ok := data.InstanceMap[l.EndInfos[0].InstanceID]
	if !ok {
		logrus.Errorf("Disable Link Between %s and %s Error: Instance %s is not in Node %d", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.EndInfos[0].InstanceID, key.NodeIndex)
		return fmt.Errorf("%s is not in node %d", l.EndInfos[0].InstanceID, key.NodeIndex)
	}
	deleteLinkReq := netreq.CreateDeleteLinkReq(l.DevInfos[0].IfIndex, instanceInfo.Pid, l.DevInfos[0].Name)
	operatorInfo.RequestChann <- &deleteLinkReq
	err := <-operatorInfo.ErrChan
	if err != nil {
		logrus.Errorf("Delete Link %s Index %d Error: %s", l.Config.LinkID, l.DevInfos[0].IfIndex, err.Error())
		return err
	}
	return nil
}

func (l *VethLink) SetParameters(para map[string]int64, operatorInfo *model.NetlinkOperatorInfo) error {
	dirtyConnect := false
	dirtyTbf := false
	dirtyNetem := false
	var instanceInfos [2]*model.Instance

	for i := 0; i < 2; i++ {
		instanceInfo, ok := data.InstanceMap[l.Config.InitEndInfos[i].InstanceID]
		if !ok {
			logrus.Errorf("Set Parameter for Link Between %s and %s Error: Instance %s is not in Node %d", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.InitEndInfos[i].InstanceID, key.NodeIndex)
			return fmt.Errorf("%s is not in node %d", l.Config.InitEndInfos[i].InstanceID, key.NodeIndex)
		}
		instanceInfos[i] = instanceInfo
	}

	for paraName, paraValue := range para {
		if paraValue == l.Parameter[paraName] {
			logrus.Debugf("Value of %s for %s is not changed, ignore.", paraName, l.Config.LinkID)
			continue
		}
		if _, ok := l.SupportParameters[paraName]; !ok {
			logrus.Warnf("Unsupport Parameter %s for Link %s.", paraName, l.Config.Type)
			continue
		}
		if paraValue != l.Parameter[paraName] {
			l.Parameter[paraName] = paraValue
			if paraName == model.ConnectParameter {
				dirtyConnect = true
			} else if paraName == VethBandwidthParameter {
				dirtyTbf = true
			} else {
				dirtyNetem = true
			}
		}
	}

	for i, v := range l.DevInfos {
		if dirtyTbf {
			tbfInfo := l.TbfQdiscTemplate
			tbfInfo.Limit = uint32(l.Parameter[VethBandwidthParameter])
			setTbfReq := netreq.CreateSetQdiscReq(
				v.IfIndex,
				instanceInfos[i].Pid,
				netreq.ReplaceQdisc,
				v.Name,
				&tbfInfo,
			)
			operatorInfo.RequestChann <- &setTbfReq
			err := <-operatorInfo.ErrChan
			if err != nil {
				logrus.Errorf("Update Tbf for Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		}

		if dirtyNetem {
			netemInfo := netlink.NewNetem(
				l.NetemQdiscTemplate.QdiscAttrs,
				netlink.NetemQdiscAttrs{
					Latency: uint32(l.Parameter[VethDelayParameter]) + 1,
					Loss:    float32(l.Parameter[VethLossParameter]) / 10000,
				},
			)

			setTbfReq := netreq.CreateSetQdiscReq(
				l.DevInfos[i].IfIndex,
				instanceInfos[i].Pid,
				netreq.ReplaceQdisc,
				l.DevInfos[i].Name,
				netemInfo,
			)
			operatorInfo.RequestChann <- &setTbfReq
			err := <-operatorInfo.ErrChan
			if err != nil {
				logrus.Errorf("Update Netem for Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		}

	}
	if dirtyConnect {
		if l.Parameter[model.ConnectParameter] == 0 {
			err := l.Disconnect(operatorInfo)
			if err != nil {
				logrus.Errorf("Disconnect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		} else {
			err := l.Connect(operatorInfo)
			if err != nil {
				logrus.Errorf("Connect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return err
			}
		}
	}
	return nil
}
