package config

import "fmt"

var (
	NodeNsKey           = ""
	NodeInstanceListKey = ""
)

func init() {
	NodeInstanceListKey = fmt.Sprintf("node_%d_instances", NodeIndex)
	NodeNsKey = fmt.Sprintf("/node_%d/ns_list", NodeIndex)
}
