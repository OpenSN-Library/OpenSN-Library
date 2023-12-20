package model

import netreq "NodeDaemon/model/netlink_request"

const ConnectParameter = "connect"

type ParameterInfo struct {
	Name           string
	MinVal         int64
	MaxVal         int64
	DefinitionFrac int64
}

var ConnectParameterInfo = ParameterInfo{
	Name:           ConnectParameter,
	MinVal:         0,
	MaxVal:         1,
	DefinitionFrac: 1,
}

type NetlinkOperatorInfo struct {
	RequestChann chan netreq.NetLinkRequest
	ErrChan      chan error
}

type Link interface {
	GetLinkConfig() LinkConfig
	GetLinkID() string
	GetLinkType() string
	Connect(operatorInfo *NetlinkOperatorInfo) error
	Disconnect(operatorInfo *NetlinkOperatorInfo) error
	Enable(operatorInfo *NetlinkOperatorInfo) error
	Disable(operatorInfo *NetlinkOperatorInfo) error
	IsCrossMachine() bool
	SetParameters(para map[string]int, operatorInfo *NetlinkOperatorInfo) error
	IsEnabled() bool
	IsConnected() bool
	GetSupportParameters() []ParameterInfo
	GetParameter(name string) (int64, error)
}

type LinkBase struct {
	Enabled           bool
	CrossMachine      bool
	SupportParameters map[string]ParameterInfo
	Parameter         map[string]int64
	Config            LinkConfig
}

func (l *LinkBase) GetLinkConfig() LinkConfig {
	return l.Config
}

func (l *LinkBase) GetLinkID() string {
	return l.Config.LinkID
}

func (l *LinkBase) GetLinkType() string {
	return l.Config.Type
}

func (l *LinkBase) IsConnected() bool {
	return l.Parameter[ConnectParameter] != 0
}

func (l *LinkBase) IsEnabled() bool {
	return l.Enabled
}

func (l *LinkBase) IsCrossMachine() bool {
	return l.CrossMachine
}

func (l *LinkBase) GetSupportParameters() map[string]ParameterInfo {
	return l.SupportParameters
}

func (l *LinkBase) GetParameter(name string) (int64, error) {
	return l.Parameter[name], nil
}
