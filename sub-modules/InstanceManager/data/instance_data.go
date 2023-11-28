package data

import "InstanceManager/model"

var Namespaces []string
var InstanceMap map[string]*model.Instance

func init() {
	InstanceMap = make(map[string]*model.Instance)
}
