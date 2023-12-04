package model

import "github.com/vishvananda/netlink"

const VethLinkType = "VethLink"

var VethLinkParameters = []ParameterInfo{
	{
		Name:           "Delay",
		MaxVal:         10000,
		MinVal:         0,
		DefinitionFrac: 1000,
	},
	{
		Name:           "Loss",
		MaxVal:         50,
		MinVal:         0,
		DefinitionFrac: 100,
	},
	{
		Name:           "Bandwidth",
		MaxVal:         10000,
		MinVal:         0,
		DefinitionFrac: 1,
	},
}

type VethLink struct {
	Config LinkConfig
	LinkID      string
	VethIfIndex [2]int
}

func (l *VethLink) GetLinkConfig() LinkConfig {
	return l.Config
}

func (l *VethLink) GetLinkID() string {
	return l.LinkID
}
func (l *VethLink) GetLinkType() string {
	return VethLinkType
}
func (l *VethLink) Connect() error {

	return nil
}
func (l *VethLink) Disconnect() error {
	link1,err := netlink.LinkByIndex(l.VethIfIndex[0])
	if err != nil {
		return nil
	}
	link2,err := netlink.LinkByIndex(l.VethIfIndex[0])
	if err != nil {
		return nil
	}
	
	return nil
}
func (l *VethLink) IsConnected() bool {

	return false
}
func (l *VethLink) GetSupportParameters() []ParameterInfo {
	return VethLinkParameters
}
func (l *VethLink) SetParameter(name string, val int64) error {

	return nil
}
func (l *VethLink) GetParameter(name string) (int64, error) {
	return 0, nil
}
