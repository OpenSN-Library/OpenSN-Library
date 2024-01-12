package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generator/ginmodel"
	"io"
	"strconv"

	"os"
)

func main() {
	var req ginmodel.CreateNamespaceReq

	req.Name = "ns1"
	req.NsConfig = ginmodel.NsReqConfig{
		ImageMap: map[string]string{
			"Satellite": "docker.io/realssd/satellite-router:latest",
		},
		ContainerEnvs: map[string]string{},
		ResourceMap: map[string]ginmodel.ResourceLimit{
			"Satellite": {
				NanoCPU:    "100M",
				MemoryByte: "32M",
			},
		},
	}
	tle_fd, err := os.Open("tle/tle.tle")
	if err != nil {
		panic(err)
	}
	buf, err := io.ReadAll(tle_fd)
	bufs := bytes.Split(buf, []byte{'\n'})
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(bufs)/3; i++ {
		newInstConfig := ginmodel.InstanceReqConfig{
			Type:               "Satellite",
			PositionChangeable: true,
			Extra: map[string]string{
				"TLE_0":          string(bufs[3*i]),
				"TLE_1":          string(bufs[3*i+1]),
				"TLE_2":          string(bufs[3*i+2]),
				"OrbitIndex":     strconv.Itoa(i / 10),
				"SatelliteIndex": strconv.Itoa(i % 10),
			},
		}
		fmt.Printf("Add Instance %d", i)
		req.InstConfigs = append(req.InstConfigs, newInstConfig)
	}
	topo_fd, err := os.Open("tle/topo.json")
	if err != nil {
		panic(err)
	}
	topo_buf, err := io.ReadAll(topo_fd)
	if err != nil {
		panic(err)
	}
	topo := make(map[string][]int)
	err = json.Unmarshal(topo_buf, &topo)
	topo_fd.Close()
	if err != nil {
		panic(err)
	}
	for k, v := range topo {
		for _, anthoer := range v {
			thisIndex, _ := strconv.Atoi(k)
			newLinkReq := ginmodel.LinkReqConfig{
				Type:          "VirtualLink",
				InstanceIndex: [2]int{thisIndex, anthoer},
				Parameter:     map[string]int64{},
			}
			fmt.Printf("Add Link Between %d and %d\n", thisIndex, anthoer)
			req.LinkConfigs = append(req.LinkConfigs, newLinkReq)
		}
	}
	tle_fd.Close()
	bytes, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	output, err := os.Create("output.json")
	if err != nil {
		panic(err)
	}
	_, err = output.Write(bytes)
	if err != nil {
		panic(err)
	}
	output.Close()
}
