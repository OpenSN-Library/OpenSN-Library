package data

import (
	"InstanceManager/config"
	"fmt"
)

var (
	NodeNsKey           = ""
	NodeInstanceListKey = ""
	NodeInstancesKey    = ""
)

func init() {
	NodeInstanceListKey = fmt.Sprintf("node_%d_instances", config.NodeIndex)
	NodeNsKey = fmt.Sprintf("/node_%d/ns_list", config.NodeIndex)
	NodeInstancesKey = fmt.Sprintf("node_%d_instances", config.NodeIndex)
}
