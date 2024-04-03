package ginmodel

import "NodeDaemon/model"

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
	NodeIndex      int       `json:"node_index"`
}

type LinkInfo struct {
	LinkAbstract
	AddressInfos [2]map[string]string `json:"address_infos"`
	EndInfos     [2]model.EndInfoType `json:"end_infos"`
}

type AddLinkEndInfo struct {
	NodeIndex  int    `json:"node_index"`
	InstanceID string `json:"instance_id"`
}

type AddLinkRequest struct {
	EndInstanceID [2]AddLinkEndInfo    `json:"end_instance_id"`
	Type          string               `json:"type"`
	InitParameter map[string]int64     `json:"init_parameter"`
	AddressInfos  [2]map[string]string `json:"address_infos"`
}
