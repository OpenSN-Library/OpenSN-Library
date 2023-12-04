package model

type ParameterInfo struct {
	Name           string
	MinVal         int64
	MaxVal         int64
	DefinitionFrac int64
}

type Link interface {
	GetLinkConfig() LinkConfig
	GetLinkID() string
	GetLinkType() string
	Connect() error
	Disconnect() error
	IsConnected() bool
	GetSupportParameters() []ParameterInfo
	SetParameter(name string, val int64) error
	GetParameter(name string) (int64, error)
}
