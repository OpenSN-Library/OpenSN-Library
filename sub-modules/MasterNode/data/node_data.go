package data

import (
	"MasterNode/model"
	"sync"
)

var NodeMap map[int]*model.Node
var NodeMapLock *sync.Mutex

func init() {

	NodeMap = make(map[int]*model.Node)
	NodeMapLock = new(sync.Mutex)
}
