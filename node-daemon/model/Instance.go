package model

const (
	SATELLITE = "Satellite"
	G_STATION = "GroundStation"
)

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type Instance struct {
	Config      InstanceConfig `json:"config"`
	ContainerID string         `json:"container_id"`
	NodeID      uint32         `json:"node_id"`
	Namespace   string         `json:"namespace"`
	LinksID     []string       `json:"links_id"`
}
