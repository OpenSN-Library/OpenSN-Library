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
	NodeInstanceListKeyTemplate       = "/node_%d/instances"
	NodeInstanceRuntimeKeyTemplate    = "/node_%d/runtime"
	NodeLinkListKeyTemplate           = "/node_%d/links"
	NodeLinkParameterKeyTemplate      = "/node_%d/link_paramter"
	NamespaceInstancePositionTemplate = "/positions/"
	NodeInstanceConfigKeyTemplate     = "/instance_config/node_%d"
)

const ( // Redis Keys
	NodeHeartBeatKey = "/node_heart_beat"
	NextNodeIndexKey = "/next_node_index"
	NextLinkIndexKey = "/next_link_index"
)

var (
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
	NodeInstanceListKeySelf = fmt.Sprintf(NodeInstanceListKeyTemplate, NodeIndex)
	NodeLinkListKeySelf = fmt.Sprintf(NodeLinkListKeyTemplate, NodeIndex)
	NodeLinkParameterKeySelf = fmt.Sprintf(NodeLinkParameterKeyTemplate, NodeIndex)
	NodeInstanceConfigKeySelf = fmt.Sprintf(NodeInstanceConfigKeyTemplate, NodeIndex)
	NodeInstanceRuntimeKeySelf = fmt.Sprintf(NodeInstanceRuntimeKeyTemplate, NodeIndex)
}
