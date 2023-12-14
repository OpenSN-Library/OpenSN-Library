package data

import (
	"NodeDaemon/model"
	"sync"
)

var NodeMap map[int]*model.Node
var NodeMapLock *sync.RWMutex

func init() {

	NodeMap = make(map[int]*model.Node)
	NodeMapLock = new(sync.RWMutex)
}
