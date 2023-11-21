package model

type Veth struct{
	Index uint32
	Name string
	MacAddr uint64
}

type Bridge struct {
	Index uint32
	Name string
}

type VethPair struct {
	VethID string
	ContainerSideVeth Veth
	KernelSideVeth Veth
	ContainerID Veth
}

type Link struct {
	LinkID string
	ConnectInstancesID [2]string
	VethIDs [2]string
	CrossNode bool
	ConnBridge *Bridge
	ConnectionChangable bool
}