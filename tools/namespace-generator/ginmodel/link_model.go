package ginmodel

type IPInfoType struct {
	V4Addr string `json:"v4_addr"`
	V6Addr string `json:"v6_addr"`
}

type LinkAbstract struct {
	LinkID    string              `json:"link_id"`
	Type      string              `json:"type"`
	Parameter map[string]int64    `json:"parameter"`
	IPInfos   [2]IPInfoType `json:"ip_infos"`
}
