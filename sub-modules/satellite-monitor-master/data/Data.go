package data

import "satellite/monitor/model"

var SatellitesInfo map[string]*model.SatelliteInfo
var IP2Satellite map[string]*model.SatelliteInfo
var GroundStationData map[string]*model.GroundStationInfo

func InitData() {
	SatellitesInfo = make(map[string]*model.SatelliteInfo)
	IP2Satellite = make(map[string]*model.SatelliteInfo)
	GroundStationData = make(map[string]*model.GroundStationInfo)
}
