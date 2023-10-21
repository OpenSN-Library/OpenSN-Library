package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"satellite/monitor/data"
	"satellite/monitor/model"
)

func RecvConn(host string, port int) {
	/**
	 * @description: receive the connection information from the satellites
	 */
	var recvObj model.ConnUdpMessage
	var buf [102400]byte
	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Println("listen on:", addr)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
	}
	for {
		length, _, err := conn.ReadFromUDP(buf[:])
		err = json.Unmarshal(buf[:length], &recvObj)
		if err != nil {
			fmt.Println(err)
		}
		if ptr, ok := data.SatellitesInfo[recvObj.NodeID]; ok {
			ptr.Open = recvObj.State
		}
	}
}

func RecvUpdate(host string, port int) {
	var buf [102400]byte
	
	addr := fmt.Sprintf("%s:%d", host, port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Println(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println(err)
	}
	for {
		var messageMap = model.UpdateMessage{}
		length, _, err := conn.ReadFromUDP(buf[:])
		fmt.Println("Recv: ",string(buf[:length]))
		err = json.Unmarshal(buf[:length], &messageMap)
		
		if err != nil {
			log.Println(err)
		}
		for k, v := range messageMap.PositionDatas {
			if _, ok := data.SatellitesInfo[k]; ok {
				data.SatellitesInfo[k].PositionInfo.Height = v.Height
				data.SatellitesInfo[k].PositionInfo.Longitude = v.Longitude
				data.SatellitesInfo[k].PositionInfo.Latitude = v.Latitude
			}
		}

		for k,v := range messageMap.GroundConnections {
			if _,ok := data.GroundStationData[k]; ok {
				data.GroundStationData[k].ConnectNodeID = v
			}
		}
	}
}
