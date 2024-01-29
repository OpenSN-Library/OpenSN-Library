package link

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/pkg/synchronizer"
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

func (l *VethLink) Connect() ([]netreq.NetLinkRequest, error) {
	if !l.Enabled {
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.LinkID)
		return nil, fmt.Errorf("%s is not enabled", l.LinkID)
	}

	var requests []netreq.NetLinkRequest

	for i, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}
		setStateReq := netreq.CreateSetStateReq(l.EndInfos[i].InstancePid, l.LinkID, true)
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
		logrus.Errorf("Connect %s and %s Error: %s is not enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.LinkID)
		return nil, fmt.Errorf("%s is not enabled", l.LinkID)
	}

	var requests []netreq.NetLinkRequest

	for _, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}
		setStateReq := netreq.CreateSetStateReq(v.InstancePid, l.LinkID, false)
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
	logrus.Infof("Enabling Link %s ,Type: Single Machine %s", l.LinkID, l.Type)
	if l.Enabled {
		logrus.Errorf("Enable %s and %s Error: %s has already been enabled", l.EndInfos[0].InstanceID, l.EndInfos[1].InstanceID, l.LinkID)
		return fmt.Errorf("%s is enabled", l.LinkID)
	}
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:      l.GetLinkID(),
			Namespace: netlink.NsPid(l.EndInfos[0].InstancePid),
		},
		PeerName:      l.GetLinkID(),
		PeerNamespace: netlink.NsPid(l.EndInfos[1].InstancePid),
	}

	err := netlink.LinkAdd(veth)

	if err != nil {
		logrus.Errorf("Add Veth Peer Link %v Error: %s", *l, err.Error())
		return err
	}

	return nil
}

func (l *VethLink) enableCrossMachine() error {
	for i, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}
		targetNodeInfo, err := synchronizer.GetNode(l.EndInfos[1-i].EndNodeIndex)
		if err != nil {
			return err
		}
		vxlanDev := netlink.Vxlan{
			LinkAttrs: netlink.LinkAttrs{
				Name:      l.LinkID,
				Namespace: netlink.NsPid(v.InstancePid),
				TxQLen:    -1,
			},
			VxlanId:  l.LinkIndex,
			SrcAddr:  key.SelfNode.L3AddrV4,
			Group:    targetNodeInfo.L3AddrV4,
			Port:     4789,
			Learning: true,
			L2miss:   true,
			L3miss:   true,
		}
		logrus.Warnf("Create Vxlan %v", vxlanDev)
		err = netlink.LinkAdd(&vxlanDev)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (l *VethLink) Enable() ([]netreq.NetLinkRequest, error) {
	if l.Enabled {
		return []netreq.NetLinkRequest{}, nil
	}
	var err error
	if l.CrossMachine {
		err = l.enableCrossMachine()
	} else {
		err = l.enableSameMachine()
	}

	if err != nil {
		return nil, err
	}
	l.Enabled = true
	return []netreq.NetLinkRequest{}, nil
}
func (l *VethLink) Disable() ([]netreq.NetLinkRequest, error) {

	if !l.Enabled {
		return []netreq.NetLinkRequest{}, nil
	}

	var requests []netreq.NetLinkRequest

	if l.CrossMachine {
		for _, v := range l.EndInfos {
			deleteLinkReq := netreq.CreateDeleteLinkReq(v.InstancePid, l.LinkID)
			requests = append(requests, &deleteLinkReq)

			return requests, nil
		}
	} else {
		deleteLinkReq := netreq.CreateDeleteLinkReq(l.EndInfos[0].InstancePid, l.LinkID)
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
	var requests []netreq.NetLinkRequest

	for paraName, paraValue := range para {
		if paraValue == l.Parameter[paraName] {
			logrus.Debugf("Value of %s for %s is not changed, ignore.", paraName, l.LinkID)
			continue
		}
		if _, ok := VirtualLinkParameterMap[paraName]; !ok {
			logrus.Warnf("Unsupport Parameter %s for Link %s.", paraName, l.Type)
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

	for _, v := range l.EndInfos {
		if v.EndNodeIndex != key.NodeIndex {
			continue
		}

		if dirtyTbf {
			tbfInfo := TbfQdiscTemplate
			tbfInfo.Limit = uint32(l.Parameter[VethBandwidthParameter])
			tbfInfo.Rate = uint64(l.Parameter[VethBandwidthParameter])
			tbfInfo.Buffer = uint32(l.Parameter[VethBandwidthParameter])
			setTbfReq := netreq.CreateSetQdiscReq(
				v.InstancePid,
				l.LinkID,
				&tbfInfo,
			)
			requests = append(requests, setTbfReq)
		}

		if dirtyNetem {
			netemInfo := netlink.NewNetem(
				NetemQdiscTemplate.QdiscAttrs,
				netlink.NetemQdiscAttrs{
					Latency: uint32(l.Parameter[VethDelayParameter]) + 1,
					Loss:    float32(l.Parameter[VethLossParameter]) / 10000,
				},
			)

			setNetemReq := netreq.CreateSetQdiscReq(
				v.InstancePid,
				l.LinkID,
				netemInfo,
			)
			requests = append(requests, setNetemReq)
		}

	}

	return requests, nil
}
