package data

import "NodeDaemon/model"

var Namespaces []string
var InstanceMap map[string]*model.Instance
var LinkMap map[string]*model.Link

func init() {
	InstanceMap = make(map[string]*model.Instance)
	LinkMap = make(map[string]*model.Link)
}
