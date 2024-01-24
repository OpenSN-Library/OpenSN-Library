package link

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/share/data"
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
			CrossMachine:      initConfig.InitEndInfos[0].EndNodeIndex != initConfig.InitEndInfos[1].EndNodeIndex,
			SupportParameters: VirtualLinkParameterMap,
			Parameter:         initConfig.InitParameter,
			Config:            initConfig,
			EndInfos:          initConfig.InitEndInfos,
		},
		DevInfos: [2]DevInfoType{
			{
				Name:    initConfig.LinkID,
				IfIndex: -1,
			},
			{
				Name:    initConfig.LinkID,
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

func (l *VethLink) Connect() ([]netreq.NetLinkRequest, error) {
	if !l.Enabled {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return nil, fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}

	var requests []netreq.NetLinkRequest

	for i, v := range l.EndInfos {
		if v.InstanceID == "" {
			logrus.Infof("Skip Link %s, because it's float link", l.Config.LinkID)
		}
		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if !ok {
			continue
		}
		setStateReq := netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, true)
		requests = append(requests, setStateReq)
	}
	logrus.Infof(
		"Connect Link %s Between %s and %s Suucess",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)
	l.Parameter[model.ConnectParameter] = 1
	return requests, nil
}
func (l *VethLink) Disconnect() ([]netreq.NetLinkRequest, error) {
	if !l.Enabled {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return nil, fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}

	var requests []netreq.NetLinkRequest

	for i, v := range l.EndInfos {

		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if !ok {
			continue
		}
		setStateReq := netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, false)
		requests = append(requests, setStateReq)
	}
	logrus.Infof(
		"Disconnect Link %s Between %s and %s Success",
		l.GetLinkID(),
		l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID,
	)
	l.Parameter[model.ConnectParameter] = 0
	return requests, nil
}

func (l *VethLink) enableSameMachine() error {
	logrus.Infof("Enabling Link %s ,Type: Single Machine %s", l.Config.LinkID, l.Config.Type)
	if l.Enabled {
		logrus.Errorf("Enable %s and %s Error: %s has already been enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID)
		return fmt.Errorf("%s is enabled", l.Config.LinkID)
	}
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:      l.DevInfos[0].Name,
			Namespace: netlink.NsPid(data.InstanceMap[l.EndInfos[0].InstanceID].Pid),
		},
		PeerName:      l.DevInfos[1].Name,
		PeerNamespace: netlink.NsPid(data.InstanceMap[l.EndInfos[1].InstanceID].Pid),
	}

	err := netlink.LinkAdd(veth)

	if err != nil {
		logrus.Errorf("Add Veth Peer Link Error: %s", err.Error())
		return err
	}

	return nil
}

func (l *VethLink) enableCrossMachine() error {
	for i, v := range l.EndInfos {
		_, ok := data.InstanceMap[v.InstanceID]
		if ok {
			vxlanDev := netlink.Vxlan{
				LinkAttrs: netlink.LinkAttrs{
					Name:      l.DevInfos[i].Name,
					Namespace: netlink.NsPid(data.InstanceMap[l.EndInfos[i].InstanceID].Pid),
					TxQLen:    -1,
				},
				VxlanId:  l.Config.LinkIndex,
				SrcAddr:  data.NodeMap[key.NodeIndex].L3AddrV4,
				Group:    data.NodeMap[l.EndInfos[1-i].EndNodeIndex].L3AddrV4,
				Port:     4789,
				Learning: true,
				L2miss:   true,
				L3miss:   true,
			}
			logrus.Warnf("Create Vxlan %v", vxlanDev)
			err := netlink.LinkAdd(&vxlanDev)
			if err != nil {
				return err
			}
			l.DevInfos[i].IfIndex = vxlanDev.Index
			return nil
		}
	}

	return nil
}

func (l *VethLink) Enable() ([]netreq.NetLinkRequest, error) {
	var err error
	if l.CrossMachine {
		err = l.enableCrossMachine()
	} else {
		err = l.enableSameMachine()
	}

	if err != nil {
		return nil, err
	}

	var requests []netreq.NetLinkRequest

	for i, v := range l.EndInfos {
		instanceInfo, ok := data.InstanceMap[v.InstanceID]
		if v.InstanceID == "" || !ok {
			continue
		}
		requests = append(
			requests,
			netreq.CreateSetStateReq(l.DevInfos[i].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name, true),
		)

		if len(l.Config.IPInfos[i].V4Addr) > 0 {
			setV4AddrReq := netreq.CreateSetV4AddrReq(
				l.DevInfos[i].IfIndex,
				instanceInfo.Pid,
				l.DevInfos[i].Name,
				l.Config.IPInfos[i].V4Addr,
			)
			requests = append(requests, setV4AddrReq)
		}
		if len(l.Config.IPInfos[i].V6Addr) > 0 {
			setV6AddrReq := netreq.CreateSetV6AddrReq(
				l.DevInfos[i].IfIndex,
				instanceInfo.Pid,
				l.DevInfos[i].Name,
				l.Config.IPInfos[i].V6Addr,
			)
			requests = append(requests, setV6AddrReq)

		}
		parameterReqs, _ := l.SetParameters(l.Config.InitParameter)
		if parameterReqs != nil {
			requests = append(requests, parameterReqs...)
		}
	}
	l.Enabled = true
	return requests, nil
}
func (l *VethLink) Disable() ([]netreq.NetLinkRequest, error) {

	if !l.Enabled {
		logrus.Errorf(
			"Disable %s and %s Error: %s is not enabled",
			l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.Config.LinkID,
		)
		return nil, fmt.Errorf("%s is not enabled", l.Config.LinkID)
	}

	var requests []netreq.NetLinkRequest

	if l.CrossMachine {
		for i, v := range l.EndInfos {
			if instanceInfo, ok := data.InstanceMap[v.InstanceID]; ok {
				deleteLinkReq := netreq.CreateDeleteLinkReq(l.DevInfos[0].IfIndex, instanceInfo.Pid, l.DevInfos[i].Name)
				requests = append(requests, &deleteLinkReq)

				return requests, nil
			}
		}
	} else {
		instanceInfo, ok := data.InstanceMap[l.EndInfos[0].InstanceID]
		if !ok {
			logrus.Errorf(
				"Disable Link Between %s and %s Error: Instance %s is not in Node %d",
				l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.EndInfos[0].InstanceID, key.NodeIndex,
			)
			return nil, fmt.Errorf("%s is not in node %d", l.EndInfos[0].InstanceID, key.NodeIndex)
		}
		deleteLinkReq := netreq.CreateDeleteLinkReq(l.DevInfos[0].IfIndex, instanceInfo.Pid, l.DevInfos[0].Name)
		requests = append(requests, &deleteLinkReq)

		return requests, nil
	}
	l.Enabled = false
	return requests, nil
}

func (l *VethLink) SetParameters(para map[string]int64) ([]netreq.NetLinkRequest, error) {
	dirtyConnect := false
	dirtyTbf := false
	dirtyNetem := false
	var instanceInfos [2]*model.Instance
	var requests []netreq.NetLinkRequest
	for i := 0; i < 2; i++ {
		instanceInfo, ok := data.InstanceMap[l.Config.InitEndInfos[i].InstanceID]
		if !ok {
			continue
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

	if dirtyConnect {
		if l.Parameter[model.ConnectParameter] == 0 {
			disconnReqs, err := l.Disconnect()
			if err != nil {
				logrus.Errorf("Disconnect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return nil, err
			}
			return disconnReqs, err
		} else {
			connReqs, err := l.Connect()
			if err != nil {
				logrus.Errorf("Connect Link Between %s and %s Error: %s", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, err.Error())
				return nil, err
			}
			requests = append(requests, connReqs...)
		}
	}

	for i, v := range l.DevInfos {

		if dirtyTbf {
			tbfInfo := l.TbfQdiscTemplate
			tbfInfo.Limit = uint32(l.Parameter[VethBandwidthParameter])
			tbfInfo.Rate = uint64(l.Parameter[VethBandwidthParameter])
			tbfInfo.Buffer = uint32(l.Parameter[VethBandwidthParameter])
			setTbfReq := netreq.CreateSetQdiscReq(
				v.IfIndex,
				instanceInfos[i].Pid,
				netreq.ReplaceQdisc,
				v.Name,
				&tbfInfo,
			)
			requests = append(requests, setTbfReq)
		}

		if dirtyNetem {
			netemInfo := netlink.NewNetem(
				l.NetemQdiscTemplate.QdiscAttrs,
				netlink.NetemQdiscAttrs{
					Latency: uint32(l.Parameter[VethDelayParameter]) + 1,
					Loss:    float32(l.Parameter[VethLossParameter]) / 10000,
				},
			)

			setNetemReq := netreq.CreateSetQdiscReq(
				l.DevInfos[i].IfIndex,
				instanceInfos[i].Pid,
				netreq.ReplaceQdisc,
				l.DevInfos[i].Name,
				netemInfo,
			)
			requests = append(requests, setNetemReq)
		}

	}

	return requests, nil
}
