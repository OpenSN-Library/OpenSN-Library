package model

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

type Link interface {
	GetLinkID() string
	GetLinkType() string
	Connect() error
	Disconnect() error
	Create() error
	Destroy() error
	Enable() error
	Disable() error
	IsCrossMachine() bool
	SetParameters(oldPara, newPara map[string]int64) error
	IsCreated() bool
	IsEnabled() bool
	SetState(newState bool)
	IsConnected() bool
	GetParameter(name string) (int64, error)
	GetEndInfos() [2]EndInfoType
	GetLinkBasePtr() *LinkBase
}

type EndInfoType struct {
	InstanceID   string `json:"instance_id"`
	InstanceType string `json:"instance_type"`
	EndNodeIndex int    `json:"end_node_index"`
}

type LinkBase struct {
	Enable       bool                 `json:"enable"`
	LinkID       string               `json:"link_id"`
	EndInfos     [2]EndInfoType       `json:"end_infos"`
	Type         string               `json:"type"`
	AddressInfos [2]map[string]string `json:"address_infos"`
	LinkIndex    int                  `json:"link_index"`
	CrossMachine bool                 `json:"cross_machine"`
	Parameter    map[string]int64     `json:"parameter"`
	NodeIndex    int                  `json:"node_index"`
}

func (l *LinkBase) SetState(newState bool) {
	l.Enable = newState
}

func (l *LinkBase) GetLinkID() string {
	return l.LinkID
}

func (l *LinkBase) GetLinkType() string {
	return l.Type
}

func (l *LinkBase) IsConnected() bool {
	return l.Parameter[ConnectParameter] != 0
}

func (l *LinkBase) IsCrossMachine() bool {
	return l.CrossMachine
}

func (l *LinkBase) GetParameter(name string) (int64, error) {
	return l.Parameter[name], nil
}

func (l *LinkBase) GetEndInfos() [2]EndInfoType {
	return l.EndInfos
}

func (l *LinkBase) GetLinkBasePtr() *LinkBase {
	return l
}
