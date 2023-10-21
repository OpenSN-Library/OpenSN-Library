package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"satellite/monitor/data"
	"satellite/monitor/model"
	"strconv"
	"time"
)

func tryGet(conn net.Conn, timeout time.Duration) ([]byte, int, error) {
	ch := make(chan []byte)
	bufLen := 0
	go func() {
		buf := make([]byte, 8192)
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			ch <- nil
		}
		bufLen = n
		ch <- buf
	}()
	select {
	case res := <-ch:
		return res, bufLen, nil
	case <-time.After(timeout):
		return nil, 0, errors.New("timeout")
	}
}

func TcpCall(targetAddr, method string, port int, args interface{}) ([]byte, int, error) {
	var payload = map[string]interface{}{
		"method": method,
		"args":   args,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", targetAddr, port))
	defer conn.Close()
	if err != nil {
		return nil, 0, err
	}

	playloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	_, err = conn.Write(playloadBytes)
	if err != nil {
		return nil, 0, err
	}
	resp, n, err := tryGet(conn, 1*time.Minute)
	return resp, n, err
}

func TracerouteTo(nodeID string, targetIP string) ([]model.TraceroutePath, error) {
	var ans []model.TraceroutePath
	var nodeRespObj model.TracerouteResp
	if info, ok := data.SatellitesInfo[nodeID]; ok {
		buf, bufLen, err := TcpCall(info.HostIP, "traceroute", 5000, []string{targetIP})
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(buf[:bufLen], &nodeRespObj)
		if err != nil {
			return nil, err
		}
		if nodeRespObj.Code != 0 {
			return nil, errors.New(nodeRespObj.Message)
		}

		for i := range nodeRespObj.Data {
			ipPart := nodeRespObj.Data[i][2]
			latency0, _ := strconv.ParseFloat(nodeRespObj.Data[i][3], 64)
			latency1, _ := strconv.ParseFloat(nodeRespObj.Data[i][5], 64)
			latency2, _ := strconv.ParseFloat(nodeRespObj.Data[i][7], 64)
			latency := (latency2 + latency1 + latency0) / 3
			ip := ipPart[1 : len(ipPart)-1]
			if node, ok := data.IP2Satellite[ip]; ok {
				if i == 0 && node.NodeID != nodeID {
					ans = append(ans, model.TraceroutePath{
						NodeID:  nodeID,
						Latency: 0,
					})
				}
				ans = append(ans, model.TraceroutePath{
					NodeID:  node.NodeID,
					Latency: latency,
				})
			}
		}
		return ans, nil
	}
	return nil, errors.New("unknown node id")
}

func StartSend(srcID,dstID1,dstID2 string) error{
	var nodeRespObj model.VideoResp
	if info, ok := data.SatellitesInfo[srcID]; ok {

		var nodeIP1,nodeIP2 string

		if getNodeIP1,ok1:=data.SatellitesInfo[dstID1];ok1 {
			nodeIP1 = getNodeIP1.TopoStartIP[0]
		} else {
			return errors.New("invalid target node ID")
		}

		if getNodeIP2,ok2:=data.SatellitesInfo[dstID2];ok2 {
			nodeIP2 = getNodeIP2.TopoStartIP[0]
		} else {
			return errors.New("invalid target node ID")
		}

		buf, bufLen, err := TcpCall(info.HostIP, "start_send", 5000, []string{info.TopoStartIP[0],nodeIP1,nodeIP2})
		if err != nil {
			return err
		}
		err = json.Unmarshal(buf[:bufLen], &nodeRespObj)
		if err != nil {
			return err
		}
		if nodeRespObj.Code != 0 {
			return errors.New(nodeRespObj.Message)
		}
	} else {
		return errors.New("invalid source node ID")
	}
	return nil
}

func StartRecv(nodeID string) error {
	var nodeRespObj model.VideoResp
	if info, ok := data.SatellitesInfo[nodeID]; ok {
		buf, bufLen, err := TcpCall(info.HostIP, "start_recv", 5000, []string{})
		if err != nil {
			return err
		}
		err = json.Unmarshal(buf[:bufLen], &nodeRespObj)
		if err != nil {
			return err
		}
		if nodeRespObj.Code != 0 {
			return errors.New(nodeRespObj.Message)
		}
	} else {
		return errors.New("invalid source node ID")
	}
	return nil
}

func StartCompress(nodeID string) error {
	var nodeRespObj model.VideoResp
	if info, ok := data.SatellitesInfo[nodeID]; ok {
		buf, bufLen, err := TcpCall(info.HostIP, "start_comp", 5000, []string{})
		if err != nil {
			return err
		}
		err = json.Unmarshal(buf[:bufLen], &nodeRespObj)
		if err != nil {
			return err
		}
		if nodeRespObj.Code != 0 {
			return errors.New(nodeRespObj.Message)
		}
	} else {
		return errors.New("invalid source node ID")
	}
	return nil
}