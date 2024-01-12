package ginmodel

type NodeAbstract struct {
	NodeID       uint32 `json:"node_id"`
	FreeInstance int    `json:"free_instance"`
	IsMasterNode bool   `json:"is_master_node"`
	L3AddrV4     string `json:"l_3_addr_v_4"`
	L3AddrV6     string `json:"l_3_addr_v_6"`
	L2Addr       string `json:"l_2_addr"` // 低六字节储存MAC地址
}

type NodeDetail struct {
	NodeAbstract
	NsInstanceMap map[string]InstanceAbstract `json:"ns_instance_map"`
	NsLinkMap     map[string]LinkAbstract     `json:"ns_link_map"`
}
