package data

import "NodeDaemon/model"

var InstanceMap map[string]*model.Instance

func init() {
	InstanceMap = make(map[string]*model.Instance)
}
