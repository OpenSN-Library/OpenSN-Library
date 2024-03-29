package ginmodel

type InstanceWebshellRequest struct {
	NodeIndex    int    `json:"node_index"`
	InstanceID   string `json:"instance_id"`
	ExpireMinute int  `json:"expire_minute"`
}

type LinkWebshellRequest struct {
	NodeIndex    int    `json:"node_index"`
	LinkID       string `json:"link_id"`
	ExpireMinute int  `json:"expire_minute"`
}
