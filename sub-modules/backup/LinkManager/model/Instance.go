package model

const (
	SATELLITE = "Satellite"
	G_STATION = "GroundStation"
)

type Position struct {
	Latitude  float64
	Longitude float64
	Altiutde  float64
}

type Instance struct {
	InstanceID        string
	Name              string
	Type              string
	PositionChangable bool
	ContainerID       string
	NodeID            uint32
	Namespace         string
	LinksID           []string
	Extra             map[string]string
}
