package data

import (
	"NodeDaemon/model"
	"sync"
)

var LinkMap map[string]model.Link
var LinkMapLock *sync.RWMutex
var TopoInfoMap map[string]*model.TopoInfo

func init() {
	LinkMap = make(map[string]model.Link)
	TopoInfoMap = make(map[string]*model.TopoInfo)
	LinkMapLock = new(sync.RWMutex)
}
