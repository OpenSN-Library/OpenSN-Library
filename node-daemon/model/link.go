package model

import netreq "NodeDaemon/model/netlink_request"

const ConnectParameter = "connect"

type ParameterInfo struct {
	Name           string `json:"name"`
	MinVal         int64  `json:"min_val"`
	MaxVal         int64  `json:"max_val"`
	DefinitionFrac int64  `json:"definition_frac"`
	DefaultVal     int64  `json:"default_val"`
}

var ConnectParameterInfo = ParameterInfo{
	Name:           ConnectParameter,
	MinVal:         0,
	MaxVal:         1,
	DefinitionFrac: 1,
}

type NetlinkOperatorInfo struct {
	RequestChann chan []netreq.NetLinkRequest
}

type Link interface {
	GetLinkConfig() LinkConfig
	GetLinkID() string
	GetLinkType() string
	Connect() ([]netreq.NetLinkRequest, error)
	Disconnect() ([]netreq.NetLinkRequest, error)
	Enable() ([]netreq.NetLinkRequest, error)
	Disable() ([]netreq.NetLinkRequest, error)
	IsCrossMachine() bool
	SetParameters(para map[string]int64) ([]netreq.NetLinkRequest, error)
	IsEnabled() bool
	IsConnected() bool
	GetParameter(name string) (int64, error)
	GetEndInfos() [2]EndInfoType
	GetLinkBasePtr() *LinkBase
}

type AddressInfoType struct {
	V4Addr string `json:"v4_addr"`
	V6Addr string `json:"v6_addr"`
}

type EndInfoType struct {
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	EndNodeIndex int    `json:"end_node_index"`
}

type LinkBase struct {
	Enabled      bool `json:"enabled"`
	CrossMachine bool `json:"cross_machine"`
	// SupportParameters map[string]ParameterInfo `json:"support_parameters"`
	Parameter map[string]int64 `json:"parameter"`
	Config    LinkConfig       `json:"config"`
	NodeIndex int              `json:"node_index"`
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

func (l *LinkBase) GetParameter(name string) (int64, error) {
	return l.Parameter[name], nil
}

func (l *LinkBase) GetEndInfos() [2]EndInfoType {
	return l.Config.EndInfos
}

func (l *LinkBase) GetLinkBasePtr() *LinkBase {
	return l
}
