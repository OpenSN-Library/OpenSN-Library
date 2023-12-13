package data

import (
	"MasterNode/model"
	"sync"
)

var NamespaceMap map[string]*model.Namespace
var NamespaceMapLock *sync.Mutex

func init() {
	NamespaceMap = make(map[string]*model.Namespace)
	NamespaceMapLock = new(sync.Mutex)
}
