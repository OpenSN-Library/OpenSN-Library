package model

type JsonResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type SetInfoReqConnection struct {
	SourceIP     string `json:"source_ip"`
	TargetIP     string `json:"target_ip"`
	TargetNodeID string `json:"target_node_id"`
}

type SetSatelliteInfoReqData struct {
	NodeID      string                 `json:"node_id"`
	HostIP      string                 `json:"host_ip"`
	Connections []SetInfoReqConnection `json:"connections"`
}

type SetSatelliteInfoReq struct {
	Total int                       `json:"total"`
	Items []SetSatelliteInfoReqData `json:"items"`
}

type LinkStateData struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
	Height    float64 `json:"height"`
	Open      bool    `json:"open"`
}

type InterfaceData struct {
	IP    string `json:"ip"`
	DstID string `json:"dst_id"`
}

type StationPositionInitData struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

type StationPositionData struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

type TransitionVideoReq struct {
	SrcID string `json:"src_id"`
	TcpDstID string `json:"tcp_dst_id"`
	MyDstID string `json:"my_dst_id"`
}