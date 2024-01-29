package link

import (
	"NodeDaemon/model"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type DevInfoType struct {
	IfIndex int    `json:"if_index"`
	Name    string `json:"name"`
}

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
	realLink, err = ParseLinkFromBase(baseLink)
	return realLink, err
}

func ParseLinkFromBase(config model.LinkBase) (model.Link, error) {
	var realLink model.Link

	switch config.Type {
	case VirtualLinkType:
		vLink := CreateVethLinkObject(config)
		realLink = vLink

	case "":
		realLink = &VethLink{}
	default:
		err := fmt.Errorf("unsupported link type: %s", config.Type)
		logrus.Errorf("Parse Link Error: %s", err.Error())
		return nil, err
	}
	realLink.GetLinkBasePtr().CrossMachine = config.EndInfos[0].EndNodeIndex != config.EndInfos[1].EndNodeIndex
	return realLink, nil
}
