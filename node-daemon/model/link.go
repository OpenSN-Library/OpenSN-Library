package model

import netreq "NodeDaemon/model/netlink_request"


type ParameterInfo struct {
	Name           string
	MinVal         int64
	MaxVal         int64
	DefinitionFrac int64
}

type NetlinkOperatorInfo struct{
	RequestChann chan netreq.NetLinkRequest
	ErrChan chan error
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
	SetParameters(para map[string]int,operatorInfo *NetlinkOperatorInfo) error
	IsEnabled() bool
	IsConnected() bool
	GetSupportParameters() []ParameterInfo
	GetParameter(name string) (int64, error)
}

type LinkBase struct {
	Enabled           bool
	Connected         bool
	CrossMachine      bool
	SupportParameters []ParameterInfo
	Parameter         map[string]int
	LinkType          string
	Config            LinkConfig
	operateChann      chan netreq.NetLinkRequest
}

func (l *LinkBase) GetLinkConfig() LinkConfig {
	return l.Config
}

func (l *LinkBase) GetLinkID() string {
	return l.Config.LinkID
}

func (l *LinkBase) GetLinkType() string {
	return l.LinkType
}

func (l *LinkBase) IsConnected() bool {
	return l.Connected
}

func (l *LinkBase) IsEnabled() bool {
	return l.Enabled
}

func (l *LinkBase) IsCrossMachine() bool {
	return l.CrossMachine
}

func (l *LinkBase) GetSupportParameters() []ParameterInfo {
	return l.SupportParameters
}

func (l *LinkBase) GetParameter(name string) (int64, error) {
	return l.Config.Parameter[name], nil
}
