package link

import (
	"NodeDaemon/model"
	netreq "NodeDaemon/model/netlink_request"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const VethLinkType = "VethLink"

type DevInfoType struct {
	IfIndex int
	Name    string
}

type EndInfoType struct {
	DevInfo    DevInfoType
	NodeIndex  int
	InstanceID string
}

type IPInfoType struct {
	V4Addr          uint32
	V6Addr          uint64
	V4Mask          uint32
	V6Mask          uint64
	SetDefaultRoute bool
}

type VethLink struct {
	model.LinkBase
	EndInfos [2]EndInfoType
	IPInfos  [2]IPInfoType
}

func (l *VethLink) Connect(operatorInfo *model.NetlinkOperatorInfo) error {

	for _, v := range l.EndInfos {
		if v.NodeIndex == key.NodeIndex {
			link, err := netlink.LinkByIndex(v.DevInfo.IfIndex)
			if err != nil {
				logrus.Errorf("Find Link By Index %d Error: %s", v.DevInfo.IfIndex, err.Error())
				return err
			}
			err = netlink.LinkSetUp(link)
			if err != nil {
				logrus.Errorf("Set Link %d Up Error: %s", v.DevInfo.IfIndex, err.Error())
				return err
			}
		}
	}

	return nil
}
func (l *VethLink) Disconnect(operatorInfo *model.NetlinkOperatorInfo) error {
	for _, v := range l.EndInfos {
		if v.NodeIndex == key.NodeIndex {
			link, err := netlink.LinkByIndex(v.DevInfo.IfIndex)
			if err != nil {
				logrus.Errorf("Find Link By Index %d Error: %s", v.DevInfo.IfIndex, err.Error())
				return err
			}
			err = netlink.LinkSetDown(link)
			if err != nil {
				logrus.Errorf("Set Link %d Up Error: %s", v.DevInfo.IfIndex, err.Error())
				return err
			}
		}
	}

	return nil
}

func (l *VethLink) Enable(operatorInfo *model.NetlinkOperatorInfo) error {

	l.EndInfos[0].DevInfo.Name = fmt.Sprintf("veth_%s_0", l.Config.LinkID)
	l.EndInfos[1].DevInfo.Name = fmt.Sprintf("veth_%s_1", l.Config.LinkID)

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = l.EndInfos[0].DevInfo.Name
	veth := &netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  l.EndInfos[1].DevInfo.Name,
	}

	err := netlink.LinkAdd(veth)

	if err != nil {
		logrus.Errorf("Add Veth Peer Link Error %s", err.Error())
		return err
	}

	l.EndInfos[0].DevInfo.IfIndex = veth.Index
	peerIndex, err := netlink.VethPeerIndex(veth)
	if err != nil {
		logrus.Errorf("Get Peer Link Index of %d Error: %s", veth.Attrs().Index, err.Error())
	}

	l.EndInfos[1].DevInfo.IfIndex = peerIndex
	initNetns, err := netns.Get()
	if err != nil {
		logrus.Errorf("Get Init Net Namespace Error: %s", err.Error())
	}
	for i, v := range l.EndInfos {
		instanceInfo := data.InstanceMap[v.InstanceID]
		setNsreq := netreq.CreateSetNetNsReq(v.DevInfo.IfIndex, int(initNetns), instanceInfo.Pid)
		operatorInfo.RequestChann <- &setNsreq
		err := <-operatorInfo.ErrChan
		if err != nil {
			return err
		}
		targetNetNs, err := netns.GetFromPid(instanceInfo.Pid)
		if err != nil {
			return err
		}
		setStateReq := netreq.CreateSetStateReq(v.DevInfo.IfIndex, int(targetNetNs), true)
		operatorInfo.RequestChann <- &setStateReq
		err = <-operatorInfo.ErrChan
		if err != nil {
			return err
		}
		setV4AddrReq := netreq.CreateSetV4AddrReq(v.DevInfo.IfIndex, int(targetNetNs), l.IPInfos[i].V4Addr, int(l.IPInfos[i].V4Mask))
		operatorInfo.RequestChann <- &setV4AddrReq
		err = <-operatorInfo.ErrChan
		if err != nil {
			return err
		}
		setV6AddrReq := netreq.CreateSetV6AddrReq(v.DevInfo.IfIndex, int(targetNetNs), l.IPInfos[i].V6Addr, int(l.IPInfos[i].V6Mask))
		operatorInfo.RequestChann <- &setV6AddrReq
		err = <-operatorInfo.ErrChan
		if err != nil {
			return err
		}
	}
	l.Enabled = true
	return nil
}
func (l *VethLink) Disable(operatorInfo *model.NetlinkOperatorInfo) error {

	return nil
}
func (l *VethLink) SetParameter(para map[string]int, operatorInfo *model.NetlinkOperatorInfo) error {

	return nil
}
