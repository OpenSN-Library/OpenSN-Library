package data

import (
	"NodeDaemon/model"
	"sync"
)

var NamespaceMap map[string]*model.Namespace
var NamespaceMapLock *sync.RWMutex

func init() {
	NamespaceMap = make(map[string]*model.Namespace)
	NamespaceMapLock = new(sync.RWMutex)
}
