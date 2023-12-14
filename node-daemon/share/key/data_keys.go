package key

import (
	"fmt"
)

var (
	NodeIndex = -1
	NodeIndexListKey = "/node_index_list"
	NodeHeartBeatKey = "node_heart_beat"
	NodesKey         = "nodes"
	NextNodeIndexKey = "next_node_index"
	NodeNsKey           = ""
	NodeInstanceListKey = ""
	NodeInstancesKey    = ""
)

func init() {
	NodeInstanceListKey = fmt.Sprintf("node_%d_instances", NodeIndex)
	NodeNsKey = fmt.Sprintf("/node_%d/ns_list", NodeIndex)
	NodeInstancesKey = fmt.Sprintf("node_%d_instances", NodeIndex)
}