package model

type TraceroutePath struct {
	NodeID  string  `json:"node_id"`
	Latency float64 `json:"latency"`
}

type SatelliteInfo struct {
	NodeID       string                    `json:"node_id"`
	Index        int                       `json:"index"`
	HostIP       string                    `json:"host_ip"`
	PositionInfo *SatellitePositionInfo    `json:"position_info"`
	Connections  map[string]ConnectionPair `json:"connections"`
	TopoStartIP  []string                  `json:"-"`
	Open         bool                      `json:"open"`
}

type ConnectionPair struct {
	SrcIP  string
	DstIp  string
	Target *SatelliteInfo
}

type SatellitePositionInfo struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Height    float64 `json:"hei"`
}

type PositionMessage struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	Height    float64 `json:"hei"`
}

type UpdateMessage struct {
	PositionDatas map[string]PositionMessage `json:"position_datas"`
	GroundConnections map[string]string `json:"ground_connections"`
}

type ConnUdpMessage struct {
	NodeID string `json:"node_id"`
	State  bool   `json:"state"`
}

type TracerouteResp struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    [][]string `json:"data"`
}

type GroundStationInfo struct {
	Latitude float64 
	Longitude float64
	ConnectNodeID string
}

type VideoResp struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    string `json:"data"`
}