package ginmodel

type GetLinkListRequest struct {
	KeyWord   string `json:"key_word"`
	PageSize  int    `json:"page_size"`
	PageIndex int    `json:"page_index"`
}

type SingleLinkRequest struct {
	NodeIndex int    `json:"node_index"`
	LinkID    string `json:"link_id"`
}

type LinkAbstract struct {
	LinkID         string    `json:"link_id"`
	Type           string    `json:"type"`
	Enable         bool      `json:"enable"`
	ConnectIntance [2]string `json:"connect_instance"`
}

type LinkInfo struct {
	LinkAbstract
	AddressInfos [2]map[string]string `json:"address_infos"`
}
