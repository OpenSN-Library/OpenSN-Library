package model

type Namespace struct {
	Name string
	AllocedInstances int
	NodeInstanceMap map[string]string
	NodeLinkMap map[string]string
}