package model

import netreq "NodeDaemon/model/netlink_request"

const ConnectParameter = "connect"

type ParameterInfo struct {
	Name           string `json:"name"`
	MinVal         int64  `json:"min_val"`
	MaxVal         int64  `json:"max_val"`
	DefinitionFrac int64  `json:"definition_frac"`
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
	SetParameters(para map[string]int64, operatorInfo *NetlinkOperatorInfo) error
	IsEnabled() bool
	IsConnected() bool
	GetSupportParameters() map[string]ParameterInfo
	GetParameter(name string) (int64, error)
}

type LinkBase struct {
	Enabled           bool                     `json:"enabled"`
	CrossMachine      bool                     `json:"cross_machine"`
	SupportParameters map[string]ParameterInfo `json:"support_parameters"`
	Parameter         map[string]int64         `json:"parameter"`
	Config            LinkConfig               `json:"config"`
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
