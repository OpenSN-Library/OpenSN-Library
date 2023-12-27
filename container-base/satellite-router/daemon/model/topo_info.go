package model

type LinkInfo struct {
	V4Addr string `json:"v4_addr"`
	V6Addr string `json:"v6_addr"`
}

type EndInfo struct {
	InstanceID string `json:"instance_id"`
	Type       string `json:"type"`
}

type TopoInfo struct {
	InstanceID string              `json:"instance_id"`
	LinkInfos  map[string]*LinkInfo `json:"link_infos"`
	EndInfos   map[string]*EndInfo  `json:"end_infos"`
}
