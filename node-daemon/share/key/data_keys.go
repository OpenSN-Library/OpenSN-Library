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
	NodeInstanceListKeyTemplate       = "/node_%d/instance_list"
	NodeLinkListKeyTemplate           = "/node_%d/link_list"
	NodeInstancesKeyTemplate          = "node_%d_instances"
	NodeLinksKeyTemplate              = "node_%d_links"
	NodeNsKeyTemplate                 = "/node_%d/ns_list"
	NodeLinkParameterKeyTemplate      = "/node_%d/link_paramter"
	NamespaceInstancePositionTemplate = "/positions/%s/"
)

const ( // Redis Keys
	NodeHeartBeatKey = "node_heart_beat"
	NodesKey         = "nodes"
	NextNodeIndexKey = "next_node_index"
	NamespacesKey    = "namespaces"
)

var (
	NodeInstancesKeySelf     = ""
	NodeLinksKeySelf         = ""
	NodeInstanceListKeySelf  = ""
	NodeLinkListKeySelf      = ""
	NodeNsKeySelf            = ""
	NodeLinkParameterKeySelf = ""
)

var (
	NodePerformanceKey = "node_performance"
	LinkPerformanceKey = "link_performance"
	InstancePerformanceKey = "instance_performance"
)

func InitKeys() {
	NodeInstancesKeySelf = fmt.Sprintf(NodeInstancesKeyTemplate, NodeIndex)
	NodeInstanceListKeySelf = fmt.Sprintf(NodeInstanceListKeyTemplate, NodeIndex)
	NodeNsKeySelf = fmt.Sprintf(NodeNsKeyTemplate, NodeIndex)
	NodeLinksKeySelf = fmt.Sprintf(NodeLinksKeyTemplate, NodeIndex)
	NodeLinkListKeySelf = fmt.Sprintf(NodeLinkListKeyTemplate, NodeIndex)
	NodeLinkParameterKeySelf = fmt.Sprintf(NodeLinkParameterKeyTemplate, NodeIndex)
}
