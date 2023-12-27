package data

import "NodeDaemon/model"

var LinkMap map[string]model.Link
var TopoInfoMap map[string]*model.TopoInfo

func init() {
	LinkMap = make(map[string]model.Link)
	TopoInfoMap = make(map[string]*model.TopoInfo)
}
