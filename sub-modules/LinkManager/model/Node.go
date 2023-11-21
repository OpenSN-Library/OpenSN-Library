
package model

const (
	MAX_INSTANCE_NODE = 128
	MASTER_NODE_MAKEUP = 16
) 

type Node struct {
	NodeID uint32
	FreeInstance int
	IsMasterNode bool
	L3AddrV4 uint32
	L3AddrV6 uint64
	L2Addr uint64 // 低六字节储存MAC地址
	NsInstanceMap map[string]string
	NsLinkMap map[string]string
}