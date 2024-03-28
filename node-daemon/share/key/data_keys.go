package key

import (
	"NodeDaemon/model"
	"fmt"
)

var (
	NodeIndex = -1
	SelfNode  *model.Node
)

const ( // Etcd Keys
	NodeIndexListKey   = "/nodes"
	EmulationConfigKey = "/emulation_config"
)

const (
	NodeWebshellRequestKeyTemplate = "/node_%d/webshell_request"
	NodeWebshellInfoKeyTemplate    = "/node_%d/webshell_info"
	NodeInstanceListKeyTemplate    = "/node_%d/instances"
	NodeInstanceRuntimeKeyTemplate = "/node_%d/runtime"
	NodeLinkListKeyTemplate        = "/node_%d/links"
	NodeLinkParameterKeyTemplate   = "/node_%d/link_parameter"
	InstancePositionKey            = "/position"
	NodeInstanceConfigKeyTemplate  = "/instance_config/node_%d"
)

const ( // Redis Keys
	NodeHeartBeatKey = "/node_heart_beat"
	NextNodeIndexKey = "/next_node_index"
	NextLinkIndexKey = "/next_link_index"
)

var (
	NodeWebshellRequestKeySelf = ""
	NodeWebshellInfoKeySelf    = ""
	NodeInstanceListKeySelf    = ""
	NodeInstanceRuntimeKeySelf = ""
	NodeLinkListKeySelf        = ""
	NodeLinkParameterKeySelf   = ""
	NodeInstanceConfigKeySelf  = ""
)

var (
	NodePerformanceKey     = "node_performance"
	LinkPerformanceKey     = "link_performance"
	InstancePerformanceKey = "instance_performance"
)

func InitKeys() {
	NodeWebshellRequestKeySelf = fmt.Sprintf(NodeWebshellRequestKeyTemplate, NodeIndex)
	NodeWebshellInfoKeySelf = fmt.Sprintf(NodeWebshellInfoKeyTemplate, NodeIndex)
	NodeInstanceListKeySelf = fmt.Sprintf(NodeInstanceListKeyTemplate, NodeIndex)
	NodeLinkListKeySelf = fmt.Sprintf(NodeLinkListKeyTemplate, NodeIndex)
	NodeLinkParameterKeySelf = fmt.Sprintf(NodeLinkParameterKeyTemplate, NodeIndex)
	NodeInstanceConfigKeySelf = fmt.Sprintf(NodeInstanceConfigKeyTemplate, NodeIndex)
	NodeInstanceRuntimeKeySelf = fmt.Sprintf(NodeInstanceRuntimeKeyTemplate, NodeIndex)
}
