package module

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/structs"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/sirupsen/logrus"
)

var captureGap = 5 * time.Second

type MonitorModule struct {
	Base
}

func CreateMonitorModule() *MonitorModule {
	return &MonitorModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			wg:         new(sync.WaitGroup),
			daemonFunc: captureNodeStatus,
			running:    false,
			ModuleName: "MonitorModule",
		},
	}
}

func uploadHostPerformanceData(
	this *model.HostResourceRaw,
	prev *model.HostResourceRaw,
	time time.Time,
) []*write.Point {
	p := influxdb2.NewPoint(
		key.NodePerformanceKey,
		map[string]string{
			"node_index": strconv.Itoa(key.NodeIndex),
		},
		structs.Map(model.HostResource{
			CPUUsage:    (this.CPUBusy - prev.CPUBusy) / (this.CPUTotal - prev.CPUTotal),
			MemByte:     (this.MemByte + prev.MemByte) / 2,
			SwapMemByte: (this.SwapMemByte + prev.SwapMemByte) / 2,
		}),
		time,
	)
	return []*write.Point{p}
}

func uploadLinkStatusInfo(
	instanceID, namespace string,
	this map[string]*model.LinkResourceRaw,
	prev map[string]*model.LinkResourceRaw,
	prevTime time.Time,
	time time.Time,
) []*write.Point {
	var points []*write.Point
	for linkID, thisInfo := range this {
		if prevInfo, ok := prev[linkID]; ok {
			points = append(points, influxdb2.NewPoint(
				key.LinkPerformanceKey,
				map[string]string{
					"node_index":  strconv.Itoa(key.NodeIndex),
					"link_id":     linkID,
					"instance_id": instanceID,
					"namespace":   namespace,
				},
				structs.Map(model.LinkResource{
					SendBps:     (float64(thisInfo.RecvByte - prevInfo.RecvByte)) / (time.Sub(prevTime).Seconds()),
					RecvBps:     (float64(thisInfo.RecvByte - prevInfo.RecvByte)) / (time.Sub(prevTime).Seconds()),
					SendPps:     float64(thisInfo.SendPack-prevInfo.SendPack) / (time.Sub(prevTime).Seconds()),
					RecvPps:     float64(thisInfo.RecvPack-prevInfo.RecvPack) / (time.Sub(prevTime).Seconds()),
					SendErrPps:  float64(thisInfo.SendErrPack-prevInfo.SendErrPack) / (time.Sub(prevTime).Seconds()),
					SendDropPps: float64(thisInfo.SendDropPack-prevInfo.SendDropPack) / (time.Sub(prevTime).Seconds()),
					RecvErrPps:  float64(thisInfo.RecvErrPack-prevInfo.RecvErrPack) / (time.Sub(prevTime).Seconds()),
					RecvDropPps: float64(thisInfo.RecvDropPack-prevInfo.RecvDropPack) / (time.Sub(prevTime).Seconds()),
				}),
				time,
			))
		}
	}
	return points
}

func uploadInstancePerformanceInfo(
	namespaceMap map[string]string,
	this map[string]*model.InstanceResouceRaw,
	prev map[string]*model.InstanceResouceRaw,
	totalTimeMs float64,
	time time.Time,
) []*write.Point {
	var points []*write.Point
	for instanceID, thisInfo := range this {
		if prevInfo, ok := prev[instanceID]; ok {
			points = append(points, influxdb2.NewPoint(
				key.InstancePerformanceKey,
				map[string]string{
					"node_index":  strconv.Itoa(key.NodeIndex),
					"instance_id": instanceID,
					"namespace":   namespaceMap[instanceID],
				},
				structs.Map(model.InstanceResouce{
					CPUUsage:    (thisInfo.CPUBusy - prevInfo.CPUBusy) / (totalTimeMs),
					MemByte:     (thisInfo.MemByte + prevInfo.MemByte) / 2,
					SwapMemByte: (thisInfo.SwapMemByte + prevInfo.SwapMemByte) / 2,
				}),
				time,
			))
		}

	}
	return points
}

func captureNodeStatus(sigChan chan int, errChan chan error) {
	var prevNodeResouce *model.HostResourceRaw
	var prevInstanceLinkResource map[string]map[string]*model.LinkResourceRaw
	var prevInstanceResource map[string]*model.InstanceResouceRaw
	var prevTime = time.Now()
	withPrev := false
	for {
		select {
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				return
			}
		case <-time.After(time.Duration(captureGap)):
			// DO NOTHING, JUST FOR NO BLOCKING
		}

		time := time.Now()
		instanceNamespaceMap := make(map[string]string)
		thisInstanceLinkResource := map[string]map[string]*model.LinkResourceRaw{}
		thisInstanceResource := map[string]*model.InstanceResouceRaw{}
		localLock := new(sync.Mutex)

		thisNodeResouce, err := utils.GetHostResourceInfo()

		if err != nil {
			logrus.Errorf("Get Host Performance Info of Node %d Error:%s", key.NodeIndex, err.Error())
			errChan <- err
		}

		instanceList, err := synchronizer.GetInstanceRuntimeList(key.NodeIndex)

		if err != nil {
			logrus.Errorf("Get Instance Runtime Info List of Node %d Error:%s", key.NodeIndex, err.Error())
			errChan <- err
		}

		wg := utils.ForEachWithThreadPool[*model.InstanceRuntime](func(instanceInfo *model.InstanceRuntime) {
			if instanceInfo.Pid == 0 {
				return
			}
			instanceResouce, err := utils.GetInstanceResourceInfo(instanceInfo.ContainerID)
			if err != nil {
				logrus.Errorf("Get Instance Performance of Instance %s Error: %s", instanceInfo.InstanceID, err.Error())
				return
			}
			linkStatus, err := utils.GetInstanceLinkResourceInfo(instanceInfo.Pid)
			if err != nil {
				logrus.Errorf("Get Link Performance of Instance %s Error: %s", instanceInfo.InstanceID, err.Error())
				return
			}
			localLock.Lock()
			thisInstanceLinkResource[instanceInfo.InstanceID] = linkStatus
			thisInstanceResource[instanceInfo.InstanceID] = instanceResouce
			localLock.Unlock()
		}, instanceList, 128)
		wg.Wait()
		if withPrev {
			points := uploadHostPerformanceData(
				thisNodeResouce,
				prevNodeResouce,
				time,
			)
			instancePoints := uploadInstancePerformanceInfo(
				instanceNamespaceMap,
				thisInstanceResource,
				prevInstanceResource,
				thisNodeResouce.CPUTotal-prevNodeResouce.CPUTotal,
				time,
			)
			for instanceID, thisLinkResouceMap := range thisInstanceLinkResource {
				if prevLinkResouceMap, ok := prevInstanceLinkResource[instanceID]; ok {
					points = append(points, uploadLinkStatusInfo(
						instanceID,
						instanceNamespaceMap[instanceID],
						thisLinkResouceMap,
						prevLinkResouceMap,
						prevTime,
						time,
					)...)
				}
			}
			points = append(points, instancePoints...)
			err = utils.InfluxDBWriteAPI.WritePoint(context.Background(), points...)
			if err != nil {
				errMsg := fmt.Sprintf("Upload Monitor Data of Node %d Error: %s", key.NodeIndex, err.Error())
				logrus.Error(errMsg)
			} else {
				logrus.Infof("Upload Monitor Data of Node %d Success.",key.NodeIndex)
			}
		} else {
			withPrev = true
		}
		prevTime = time
		prevInstanceLinkResource = thisInstanceLinkResource
		prevInstanceResource = thisInstanceResource
		prevNodeResouce = thisNodeResouce
	}
}
