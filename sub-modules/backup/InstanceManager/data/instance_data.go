package data

import "InstanceManager/model"

var NamespacesMap map[string]string
var InstanceMap map[string]*model.Instance

func init() {
	InstanceMap = make(map[string]*model.Instance)
}
