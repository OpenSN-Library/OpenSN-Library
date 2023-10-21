package handler

import (
	"fmt"
	"net/http"
	"satellite/monitor/data"
	"satellite/monitor/model"
	"satellite/monitor/utils"

	"github.com/gin-gonic/gin"
)

func GetConnectionInfo(context *gin.Context) {

	ans := map[string][]string{}

	for k, v := range data.SatellitesInfo {
		for _, connectKey := range v.TopoStartIP {
			if arr, ok := ans[k]; ok {
				ans[k] = append(arr, v.Connections[connectKey].Target.NodeID)
			} else {
				ans[k] = []string{v.Connections[connectKey].Target.NodeID}
			}
		}
	}

	resp := model.JsonResp{
		Code:    0,
		Message: "success",
		Data:    ans,
	}
	context.JSON(http.StatusOK, resp)
	return
}

func SetConnectionInfo(context *gin.Context) {
	var reqBody model.SetSatelliteInfoReq
	resp := model.JsonResp{
		Code:    0,
		Message: "success",
	}
	err := context.BindJSON(&reqBody)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		context.JSON(http.StatusBadRequest, resp)
		return
	}

	for i, v := range reqBody.Items {
		satelliteInfo := new(model.SatelliteInfo)
		satelliteInfo.NodeID = v.NodeID
		satelliteInfo.Index = i
		satelliteInfo.HostIP = v.HostIP
		satelliteInfo.Connections = make(map[string]model.ConnectionPair)
		satelliteInfo.PositionInfo = new(model.SatellitePositionInfo)
		// data.satellitesInfo is used to store the information of the satellites
		if _, ok := data.SatellitesInfo[v.NodeID]; ok {
			continue
		}
		data.SatellitesInfo[v.NodeID] = satelliteInfo
	}

	for _, v := range reqBody.Items {
		for _, c := range v.Connections {
			srcNodeID := v.NodeID
			dstNodeID := c.TargetNodeID
			if srcSatellite, ok := data.SatellitesInfo[srcNodeID]; ok {
				srcSatellite.Connections[c.SourceIP] = model.ConnectionPair{
					SrcIP:  c.SourceIP,
					DstIp:  c.TargetIP,
					Target: data.SatellitesInfo[c.TargetNodeID],
				}
				srcSatellite.TopoStartIP = append(srcSatellite.TopoStartIP, c.SourceIP)
				data.IP2Satellite[c.TargetIP] = data.SatellitesInfo[c.TargetNodeID]
			}
			if dstSatellite, ok := data.SatellitesInfo[dstNodeID]; ok {
				dstSatellite.Connections[c.SourceIP] = model.ConnectionPair{
					SrcIP:  c.TargetIP,
					DstIp:  c.SourceIP,
					Target: data.SatellitesInfo[v.NodeID],
				}
				data.IP2Satellite[c.SourceIP] = data.SatellitesInfo[v.NodeID]
			}
		}
	}

	// traverse the data.SatellitesInfo and print
	for _, v := range data.SatellitesInfo {
		fmt.Println("-------------------------------------------")
		fmt.Println("Container connect to docker via: ", v.HostIP)
		fmt.Println("Container ID: ", v.NodeID)
		fmt.Println("Container index: ", v.Index)
		fmt.Println("Satellite state: ", v.Open)
		fmt.Println("Container connections: ")
		for _, c := range v.Connections {
			fmt.Println("    ", c.SrcIP, " -> ", c.DstIp, " of ", c.Target.NodeID)
		}
		fmt.Println("-------------------------------------------")
	}

	context.JSON(http.StatusOK, resp)
	return
}


func PrintHandler(context *gin.Context){
	resp := model.JsonResp{
		Code:    0,
		Message: "success",
	}
	// traverse the data.SatellitesInfo and print
	for _, v := range data.SatellitesInfo {
		fmt.Println("-------------------------------------------")
		fmt.Println("Container connect to docker via: ", v.HostIP)
		fmt.Println("Container ID: ", v.NodeID)
		fmt.Println("Container index: ", v.Index)
		fmt.Println("Satellite state: ", v.Open)
		fmt.Println("Container connections: ")
		for _, c := range v.Connections {
			fmt.Println("    ", c.SrcIP, " -> ", c.DstIp, " of ", c.Target.NodeID)
		}
		fmt.Println("Satellite position: ")
		fmt.Println("    Latitude: ", v.PositionInfo.Latitude)
		fmt.Println("    Longitude: ", v.PositionInfo.Longitude)
		fmt.Println("    Height: ", v.PositionInfo.Height)
		fmt.Println("-------------------------------------------")
	}
	context.JSON(http.StatusOK, resp)
	return
}

func GetLinkState(context *gin.Context) {

	var respData = map[string]model.LinkStateData{}

	for k, v := range data.SatellitesInfo {
		respData[k] = model.LinkStateData{
			Latitude:  v.PositionInfo.Latitude,
			Longitude: v.PositionInfo.Longitude,
			Height:    v.PositionInfo.Height,
			Open:      v.Open,
		}
	}

	resp := model.JsonResp{
		Code:    0,
		Message: "success",
		Data:    respData,
	}

	context.JSON(http.StatusOK, resp)
	return
}

func GetTracerouteResultHandler(context *gin.Context) {
	var resp model.JsonResp
	dstIP := context.DefaultQuery("dst_ip", "")
	srcID := context.DefaultQuery("src_id", "")
	if dstIP == "" || srcID == "" {
		resp.Code = -1
		resp.Message = "Invalid Parameter."
		context.JSON(http.StatusBadRequest, resp)
		return
	}
	result, err := utils.TracerouteTo(srcID, dstIP)
	if err != nil {
		resp.Code = -1
		resp.Message = err.Error()
		context.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp.Code = 0
	resp.Message = "success"
	resp.Data = result
	context.JSON(http.StatusOK, resp)
	return
}

func GetInterfacesHandler(context *gin.Context) {
	var resp model.JsonResp
	var result = map[string][]model.InterfaceData{}
	for nodeID, info := range data.SatellitesInfo {
		var array []model.InterfaceData
		for _, v := range info.Connections {
			array = append(array, model.InterfaceData{
				IP:    v.SrcIP,
				DstID: v.Target.NodeID,
			})
		}
		result[nodeID] = array
	}
	resp.Code = 0
	resp.Message = "success"
	resp.Data = result
	context.JSON(http.StatusOK, resp)
	return
}

func SetGroundStationPositionHandler(ctx *gin.Context) {
	var updateData map[string]model.StationPositionInitData
	ctx.ShouldBindJSON(&updateData)
	for k,v := range updateData {
		data.GroundStationData[k] = &model.GroundStationInfo{
			Latitude: v.Latitude,
			Longitude: v.Longitude,
		}
	}
	jsonResp := model.JsonResp {
		Code: 0,
		Message: "success",
		Data: nil,
	}
	ctx.JSON(http.StatusOK,jsonResp)
}

func GetGroundStationPositionHandler(ctx *gin.Context) {
	respData := map[string]model.StationPositionData{}
	for k,v := range data.GroundStationData {
		respData[k] = model.StationPositionData{
			Latitude: v.Latitude,
			Longitude: v.Longitude,
		} 
		
	}
	jsonResp := model.JsonResp {
		Code: 0,
		Message: "success",
		Data: respData,
	}
	ctx.JSON(http.StatusOK,jsonResp)
}

func GetGroundStationConnHandler(ctx *gin.Context) {
	respData := map[string]string{}
	for k,v := range data.GroundStationData {
		respData[k] = v.ConnectNodeID
	}
	jsonResp := model.JsonResp {
		Code: 0,
		Message: "success",
		Data: respData,
	}
	ctx.JSON(http.StatusOK,jsonResp)
}