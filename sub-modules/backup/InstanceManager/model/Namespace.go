package model

type Namespace struct {
	Name               string
	AllocatedInstances int
	NodeInstanceMap    map[string]string
	NodeLinkMap        map[string]string
}
