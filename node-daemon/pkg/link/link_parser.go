package link

import (
	"NodeDaemon/model"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

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

func ParseLinkFromBytes(seq []byte) (model.Link, error) {
	var baseLink model.LinkBase
	var realLink model.Link
	err := json.Unmarshal(seq, &baseLink)
	if err != nil {
		logrus.Error("Unmarshal Json Data to Link Base Error, Redis Data May Crash: ", err.Error())
		return nil, err
	}

	switch baseLink.Config.Type {
	case VirtualLinkType:
		vLink := new(VethLink)
		err := json.Unmarshal(seq, &vLink)
		if err != nil {
			logrus.Error("Unmarshal Json Data to Veth Link Error, Redis Data May Crash: ", err.Error())
			return nil, err
		}
		realLink = vLink
	default:
		err := fmt.Errorf("unsupported link type: %s", baseLink.GetLinkType())
		logrus.Errorf("Parse Link Error: %s", err.Error())
		return nil, err
	}
	return realLink, nil
}

func ParseLinkFromConfig(config model.LinkConfig,nodeIndex int) (model.Link, error) {
	var realLink model.Link

	switch config.Type {
	case VirtualLinkType:
		vLink := CreateVethLinkObject(config)
		vLink.NodeIndex = nodeIndex
		realLink = vLink
	default:
		err := fmt.Errorf("unsupported link type: %s", config.Type)
		logrus.Errorf("Parse Link Error: %s", err.Error())
		return nil, err
	}
	
	return realLink, nil
}
