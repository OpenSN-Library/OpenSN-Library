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

const FixPhysicalLinkType = "FixPhysicalLink"

const (
	FixPhysicalDelayParameter     = "delay"
	FixPhysicalLossParameter      = "loss"
	FixPhysicalBandwidthParameter = "bandwidth"
)

var FixPhysicalLinkParameterMap = map[string]model.ParameterInfo{
	model.ConnectParameter: model.ConnectParameterInfo,
	FixPhysicalDelayParameter: {
		Name:           FixPhysicalDelayParameter,
		MinVal:         0,
		MaxVal:         1e10,
		DefinitionFrac: 1e9,
		DefaultVal:     0,
	},
	FixPhysicalLossParameter: {
		Name:           FixPhysicalLossParameter,
		MinVal:         0,
		MaxVal:         10000,
		DefinitionFrac: 10000,
		DefaultVal:     0,
	},
	FixPhysicalBandwidthParameter: {
		Name:           FixPhysicalBandwidthParameter,
		MinVal:         0,
		MaxVal:         1e10,
		DefinitionFrac: 1,
		DefaultVal:     0,
	},
}

type FixPLinkCrossMachineDev struct {
	bridgeIndex int
}

type FixPhysicalLink struct {
	model.LinkBase
	crossMachineDev  FixPLinkCrossMachineDev
	DevInfos         [2]DevInfoType `json:"dev_info"`
	TbfQdiscTemplate netlink.Tbf
}

func CreateFixPhysicalLinkObject(initConfig model.LinkConfig) *FixPhysicalLink {

	return &FixPhysicalLink{
		LinkBase: model.LinkBase{
			Enabled:           false,
			CrossMachine:      false,
			SupportParameters: FixPhysicalLinkParameterMap,
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
		TbfQdiscTemplate: netlink.Tbf{
			QdiscAttrs: netlink.QdiscAttrs{
				Handle: netlink.MakeHandle(1, 0),
				Parent: netlink.HANDLE_ROOT,
			},
		},
	}
}

func (l *FixPhysicalLink) Connect(operatorInfo *model.NetlinkOperatorInfo) error {
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
func (l *FixPhysicalLink) Disconnect(operatorInfo *model.NetlinkOperatorInfo) error {
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

func (l *FixPhysicalLink) enableSameMachine() error {

	return nil
}

func (l *FixPhysicalLink) enableCrossMachine() error {

	return nil
}

func (l *FixPhysicalLink) Enable(operatorInfo *model.NetlinkOperatorInfo) error {
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
func (l *FixPhysicalLink) Disable(operatorInfo *model.NetlinkOperatorInfo) error {
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

func (l *FixPhysicalLink) SetParameters(para map[string]int64, operatorInfo *model.NetlinkOperatorInfo) error {
	dirtyConnect := false
	dirtyTbf := false
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
			} else if paraName == FixPhysicalBandwidthParameter {
				dirtyTbf = true
			}
		}
	}

	for i, v := range l.DevInfos {
		if dirtyTbf {
			tbfInfo := l.TbfQdiscTemplate
			tbfInfo.Limit = uint32(l.Parameter[FixPhysicalBandwidthParameter])
			tbfInfo.Rate = uint64(l.Parameter[FixPhysicalBandwidthParameter])
			tbfInfo.Buffer = uint32(l.Parameter[FixPhysicalBandwidthParameter])
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
