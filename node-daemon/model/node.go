package model

type Node struct {
	NodeID             uint32            `json:"node_id"`
	FreeInstance       int               `json:"free_instance"`
	IsMasterNode       bool              `json:"is_master_node"`
	L3AddrV4           []byte            `json:"l_3_addr_v_4"`
	L3AddrV6           []byte            `json:"l_3_addr_v_6"`
	L2Addr             []byte            `json:"l_2_addr"` // 低六字节储存MAC地址
	NsInstanceMap      map[string]string `json:"ns_instance_map"`
	NsLinkMap          map[string]string `json:"ns_link_map"`
	NodeLinkDeviceInfo map[string]int    `json:"node_link_device_info"`
}
