package data

import (
	"MasterNode/model"
	"sync"
)

var NodeMap map[int]*model.Node
var NodeMapLock *sync.Mutex

var NamespaceMap map[string]*model.Namespace
var NamespaceMapLock *sync.Mutex

func init() {
	NamespaceMap = make(map[string]*model.Namespace)
	NamespaceMapLock = new(sync.Mutex)
	NodeMap = make(map[int]*model.Node)
	NodeMapLock = new(sync.Mutex)
}