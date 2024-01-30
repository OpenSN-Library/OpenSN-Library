package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generator/ginmodel"
	"io"
	"os"
	"strconv"
)

func main() {
	var emuConfig ginmodel.ConfigEmulationReq = ginmodel.ConfigEmulationReq{
		"Satellite": ginmodel.InstanceTypeConfig{
			Image: "docker.io/realssd/satellite-router:latest",
			ResourceLimit: ginmodel.ResourceLimit{
				NanoCPU:    "10M",
				MemoryByte: "24M",
			},
		},
		"GroundStation": ginmodel.InstanceTypeConfig{
			Image: "docker.io/realssd/satellite-router:latest",
			ResourceLimit: ginmodel.ResourceLimit{
				NanoCPU:    "10M",
				MemoryByte: "24M",
			},
		},
	}

	emuBytes, err := json.Marshal(emuConfig)
	if err != nil {
		panic(err)
	}
	output, err := os.Create("EmulationConfigRequest.json")
	if err != nil {
		panic(err)
	}
	_, err = output.Write(emuBytes)
	if err != nil {
		panic(err)
	}
	output.Close()

	tle_fd, err := os.Open("tle/tle.tle")
	if err != nil {
		panic(err)
	}

	var topoReq ginmodel.AddTopologyReq

	buf, err := io.ReadAll(tle_fd)
	bufs := bytes.Split(buf, []byte{'\n'})
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(bufs)/3; i++ {
		newInstConfig := ginmodel.TopologyInstance{
			Type: "Satellite",
			Extra: map[string]string{
				"TLE_0":          string(bufs[3*i]),
				"TLE_1":          string(bufs[3*i+1]),
				"TLE_2":          string(bufs[3*i+2]),
				"OrbitIndex":     strconv.Itoa(i / 10),
				"SatelliteIndex": strconv.Itoa(i % 10),
			},
		}
		fmt.Printf("Add Instance %d", i)
		topoReq.Instances = append(topoReq.Instances, newInstConfig)
	}

	for i := 0; i < 10; i++ {
		newInstConfig := ginmodel.TopologyInstance{
			Type: "GroundStation",
			Extra: map[string]string{
				"latitude":     strconv.FormatFloat(0.1*float64(i), 'f', 6, 64),
				"longitude":    strconv.FormatFloat(0.1*float64(i), 'f', 6, 64),
				"altitude":     "0",
				"ground_index": strconv.Itoa(i),
			},
		}
		topoReq.Instances = append(topoReq.Instances, newInstConfig)
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
			newLinkReq := ginmodel.TopologyLink{
				Type:       "vlink",
				EndIndexes: [2]int{thisIndex, anthoer},
			}
			fmt.Printf("Add Link Between %d and %d\n", thisIndex, anthoer)
			topoReq.Links = append(topoReq.Links, newLinkReq)
		}
	}
	tle_fd.Close()
	bytes, err := json.Marshal(topoReq)
	if err != nil {
		panic(err)
	}
	output, err = os.Create("InitTopology.json")
	if err != nil {
		panic(err)
	}
	_, err = output.Write(bytes)
	if err != nil {
		panic(err)
	}
	output.Close()
}
