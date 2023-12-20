package data

import "NodeDaemon/model"

var LinkMap map[string]*model.Link

func init() {
	LinkMap = make(map[string]*model.Link)
}
