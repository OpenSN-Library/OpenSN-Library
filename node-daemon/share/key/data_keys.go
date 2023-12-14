package key

import (
	"fmt"
)

var (
	NodeIndex = -1
)

const ( // Etcd Keys
	NodeIndexListKey = "/node_index_list"
)

const (
	NodeInstanceListKeyTemplate = "/node_%d/instance_list"
	NodeInstancesKeyTemplate    = "node_%d_instances"
	NodeNsKeyTemplate           = "/node_%d/ns_list"
)

const ( // Redis Keys
	NodeHeartBeatKey = "node_heart_beat"
	NodesKey         = "nodes"
	NextNodeIndexKey = "next_node_index"
	NamespacesKey    = "namespaces"
)

var (
	NodeInstancesKeySelf    = ""
	NodeInstanceListKeySelf = ""
	NodeNsKeySelf           = ""
)

func InitKeys() {
	NodeInstancesKeySelf = fmt.Sprintf(NodeInstancesKeyTemplate, NodeIndex)
	NodeInstanceListKeySelf = fmt.Sprintf(NodeInstanceListKeyTemplate, NodeIndex)
	NodeNsKeySelf = fmt.Sprintf(NodeNsKeyTemplate, NodeIndex)
}
