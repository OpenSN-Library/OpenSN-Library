package config

import "fmt"

var (
	NodeNsKey           = ""
	NodeInstanceListKey = ""
)

func init() {
	NodeNsKey = fmt.Sprintf("node_%d_instances", NodeIndex)
	NodeNsKey = fmt.Sprintf("/node_%d/ns_list", NodeIndex)
}
