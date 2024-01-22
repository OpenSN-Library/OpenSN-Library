package ginmodel

import "NodeDaemon/model"

type LinkAbstract struct {
	LinkID         string              `json:"link_id"`
	Type           string              `json:"type"`
	Parameter      map[string]int64    `json:"parameter"`
	IPInfos        [2]model.IPInfoType `json:"ip_infos"`
	ConnectIntance [2]string           `json:"connect_instance"`
}
