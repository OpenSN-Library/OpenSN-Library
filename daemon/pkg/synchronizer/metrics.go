package synchronizer

import (
	"NodeDaemon/config"
	"NodeDaemon/model"
	"NodeDaemon/utils"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func GetLastNodeResourceDatas(nodeIndexes []int) ([]model.HostResource, error) {
	var hostResources []model.HostResource
	for _, nodeIndex := range nodeIndexes {
		hostResource := model.HostResource{}
		tableResult, err := utils.InfluxDBQueryAPI.Query(
			context.Background(),
			fmt.Sprintf(
				`
				from(bucket:"%s") |>
				range(start: -1m) |>
				filter(fn: (r) => r._measurement == "node_performance" and r.node_index == "%d") |>
				top(n:1, columns: ["_time"]) 
				`, config.GlobalConfig.Dependency.InfluxdbBucket, nodeIndex,
			),
		)
		if err != nil {
			logrus.Errorf("Get Last Node Resource Data Error: %s", err.Error())
			hostResources = append(hostResources, hostResource)
			continue
		}
		for tableResult.Next() {
			switch tableResult.Record().Field() {
			case "CPUUsage":
				hostResource.CPUUsage = tableResult.Record().Value().(float64)
				hostResource.Time = tableResult.Record().Time()
			case "MemByte":
				hostResource.MemByte = tableResult.Record().Value().(uint64)
			case "SwapMemByte":
				hostResource.SwapMemByte = tableResult.Record().Value().(uint64)
			default:
				logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
			}
		}
		hostResources = append(hostResources, hostResource)
	}
	return hostResources, nil
}

func GetPeriodNodeResourceDatas(periodExpr string, nodeIndexes []int) ([][]*model.HostResource, error) {
	var hostResources [][]*model.HostResource
	for _, nodeIndex := range nodeIndexes {
		hostResource := []*model.HostResource{}
		tableResult, err := utils.InfluxDBQueryAPI.Query(
			context.Background(),
			fmt.Sprintf(
				`
				from(bucket:"%s") |>
				range(start: -%s) |>
				filter(fn: (r) => r._measurement == "node_performance" and r.node_index == "%d") 
				`, config.GlobalConfig.Dependency.InfluxdbBucket, periodExpr, nodeIndex,
			),
		)
		if err != nil {
			logrus.Errorf("Get Period Node Resource Data Error: %s", err.Error())
			hostResources = append(hostResources, hostResource)
			continue
		}
		for tableResult.Next() {
			var target *model.HostResource
			for i := len(hostResource) - 1; i >= 0; i-- {
				if hostResource[i].Time == tableResult.Record().Time() {
					target = hostResource[i]
					break
				}
			}
			if target == nil {
				target = &model.HostResource{}
				hostResource = append(hostResource, target)
			}
			switch tableResult.Record().Field() {
			case "CPUUsage":
				target.CPUUsage = tableResult.Record().Value().(float64)
				target.Time = tableResult.Record().Time()
			case "MemByte":
				target.MemByte = tableResult.Record().Value().(uint64)
			case "SwapMemByte":
				target.SwapMemByte = tableResult.Record().Value().(uint64)

			default:
				logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
			}
		}
		hostResources = append(hostResources, hostResource)
	}
	return hostResources, nil
}

func GetLastInstanceResourceDatas(instanceIDs []string) ([]model.InstanceResouce, error) {
	var instanceResources []model.InstanceResouce
	for _, instanceID := range instanceIDs {
		instanceResource := model.InstanceResouce{}
		tableResult, err := utils.InfluxDBQueryAPI.Query(
			context.Background(),
			fmt.Sprintf(
				`
				from(bucket:"%s") |>
				range(start: -1m) |>
				filter(fn: (r) => r._measurement == "instance_performance" and r.instance_id == "%s") |>
				top(n:1, columns: ["_time"])
				`, config.GlobalConfig.Dependency.InfluxdbBucket, instanceID,
			),
		)
		if err != nil {
			logrus.Errorf("Get Last Node Resource Data Error: %s", err.Error())
			instanceResources = append(instanceResources, instanceResource)
			continue
		}
		for tableResult.Next() {
			switch tableResult.Record().Field() {
			case "CPUUsage":
				instanceResource.CPUUsage = tableResult.Record().Value().(float64)
				instanceResource.Time = tableResult.Record().Time()
			case "MemByte":
				instanceResource.MemByte = tableResult.Record().Value().(uint64)
			case "SwapMemByte":
				instanceResource.SwapMemByte = tableResult.Record().Value().(uint64)
			default:
				logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
			}
		}
		instanceResources = append(instanceResources, instanceResource)
	}
	return instanceResources, nil
}

func GetLastAllInstanceResourceDatas() (map[string]*model.InstanceResouce, error) {
	instanceResources := map[string]*model.InstanceResouce{}

	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
				from(bucket:"%s") |>
				range(start: -1m) |>
				filter(fn: (r) => r._measurement == "instance_performance") |>
				top(n:1, columns: ["_time"])
				`, config.GlobalConfig.Dependency.InfluxdbBucket,
		),
	)
	if err != nil {
		logrus.Errorf("Get Last All Instance Metrics Data Error: %s", err.Error())
		return instanceResources, err
	}
	for tableResult.Next() {
		instanceID := tableResult.Record().ValueByKey("instance_id").(string)
		if _, ok := instanceResources[instanceID]; !ok {
			instanceResources[instanceID] = &model.InstanceResouce{}
		}
		switch tableResult.Record().Field() {
		case "CPUUsage":
			instanceResources[instanceID].CPUUsage = tableResult.Record().Value().(float64)
			instanceResources[instanceID].Time = tableResult.Record().Time()
		case "MemByte":
			instanceResources[instanceID].MemByte = tableResult.Record().Value().(uint64)
		case "SwapMemByte":
			instanceResources[instanceID].SwapMemByte = tableResult.Record().Value().(uint64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return instanceResources, nil
}

func GetPeriodInstanceResourceDatas(periodExpr string, instanceID string) ([]*model.InstanceResouce, error) {
	var instanceResources []*model.InstanceResouce
	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
			from(bucket:"%s") |>
			range(start: -%s) |>
			filter(fn: (r) => r._measurement == "instance_performance" and r.instance_id == "%s") 
			`, config.GlobalConfig.Dependency.InfluxdbBucket, periodExpr, instanceID,
		),
	)
	if err != nil {
		logrus.Errorf("Get Period Instance Resource Data Error: %s", err.Error())
		return instanceResources, err
	}
	for tableResult.Next() {
		var target *model.InstanceResouce
		for i := len(instanceResources) - 1; i >= 0; i-- {
			if instanceResources[i].Time == tableResult.Record().Time() {
				target = instanceResources[i]
				break
			}
		}
		if target == nil {
			target = &model.InstanceResouce{}
			instanceResources = append(instanceResources, target)
		}
		switch tableResult.Record().Field() {
		case "CPUUsage":
			target.CPUUsage = tableResult.Record().Value().(float64)
			target.Time = tableResult.Record().Time()
		case "MemByte":
			target.MemByte = tableResult.Record().Value().(uint64)
		case "SwapMemByte":
			target.SwapMemByte = tableResult.Record().Value().(uint64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return instanceResources, nil
}

func GetPeriodAllInstanceResourceDatas(periodExpr string) (map[string][]*model.InstanceResouce, error) {
	instanceResources := map[string][]*model.InstanceResouce{}

	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
				from(bucket:"%s") |>
				range(start: -%s) |>
				filter(fn: (r) => r._measurement == "instance_performance") 
				`, config.GlobalConfig.Dependency.InfluxdbBucket, periodExpr,
		),
	)
	if err != nil {
		logrus.Errorf("Get Period All Instance Metrics Data Error: %s", err.Error())
		return instanceResources, err
	}
	for tableResult.Next() {
		instanceID := tableResult.Record().ValueByKey("instance_id").(string)
		if _, ok := instanceResources[instanceID]; !ok {
			instanceResources[instanceID] = []*model.InstanceResouce{}
		}
		var target *model.InstanceResouce
		for i := len(instanceResources[instanceID]) - 1; i >= 0; i-- {
			if instanceResources[instanceID][i].Time == tableResult.Record().Time() {
				target = instanceResources[instanceID][i]
				break
			}
		}
		if target == nil {
			target = &model.InstanceResouce{}
			instanceResources[instanceID] = append(instanceResources[instanceID], target)
		}
		switch tableResult.Record().Field() {
		case "CPUUsage":
			target.CPUUsage = tableResult.Record().Value().(float64)
			target.Time = tableResult.Record().Time()
		case "MemByte":
			target.MemByte = tableResult.Record().Value().(uint64)
		case "SwapMemByte":
			target.SwapMemByte = tableResult.Record().Value().(uint64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return instanceResources, nil
}

func GetLastLinkResourceDatas(linkIDs []string) ([]model.LinkResource, error) {
	var linkResources []model.LinkResource
	for _, linkID := range linkIDs {
		linkResource := model.LinkResource{}
		tableResult, err := utils.InfluxDBQueryAPI.Query(
			context.Background(),
			fmt.Sprintf(
				`
				from(bucket:"%s") |>
				range(start: -1m) |>
				filter(fn: (r) => r._measurement == "link_performance" and r.link_id == "%s") |>
				top(n:1, columns: ["_time"])
				`, config.GlobalConfig.Dependency.InfluxdbBucket, linkID,
			),
		)
		if err != nil {
			logrus.Errorf("Get Last Link Resource Data of %s Error: %s", linkID, err.Error())
			linkResources = append(linkResources, linkResource)
			continue
		}
		for tableResult.Next() {
			switch tableResult.Record().Field() {
			case "SendBps":
				linkResource.SendBps += tableResult.Record().Value().(float64)
				linkResource.Time = tableResult.Record().Time()
			case "RecvBps":
				linkResource.RecvBps += tableResult.Record().Value().(float64)
			case "SendPps":
				linkResource.SendPps += tableResult.Record().Value().(float64)
			case "RecvPps":
				linkResource.RecvPps += tableResult.Record().Value().(float64)
			case "SendErrPps":
				linkResource.SendErrPps += tableResult.Record().Value().(float64)
			case "SendDropPps":
				linkResource.SendDropPps += tableResult.Record().Value().(float64)
			case "RecvErrPps":
				linkResource.RecvErrPps += tableResult.Record().Value().(float64)
			case "RecvDropPps":
				linkResource.RecvDropPps += tableResult.Record().Value().(float64)
			default:
				logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
			}
		}
		linkResources = append(linkResources, linkResource)
	}
	return linkResources, nil
}

func GetLastAllLinkResourceDatas() (map[string]*model.LinkResource, error) {
	linkResources := map[string]*model.LinkResource{}

	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
				from(bucket:"%s") |>
				range(start: -1m) |>
				filter(fn: (r) => r._measurement == "link_performance") |>
				top(n:1, columns: ["_time"])
				`, config.GlobalConfig.Dependency.InfluxdbBucket,
		),
	)
	if err != nil {
		logrus.Errorf("Get Last All  Link Metrics Data Error: %s", err.Error())
		return linkResources, err
	}
	for tableResult.Next() {
		linkID := tableResult.Record().ValueByKey("link_id").(string)
		if _, ok := linkResources[linkID]; !ok {
			linkResources[linkID] = &model.LinkResource{}
		}
		switch tableResult.Record().Field() {

		case "SendBps":
			linkResources[linkID].SendBps += tableResult.Record().Value().(float64)
			linkResources[linkID].Time = tableResult.Record().Time()
		case "RecvBps":
			linkResources[linkID].RecvBps += tableResult.Record().Value().(float64)
		case "SendPps":
			linkResources[linkID].SendPps += tableResult.Record().Value().(float64)
		case "RecvPps":
			linkResources[linkID].RecvPps += tableResult.Record().Value().(float64)
		case "SendErrPps":
			linkResources[linkID].SendErrPps += tableResult.Record().Value().(float64)
		case "SendDropPps":
			linkResources[linkID].SendDropPps += tableResult.Record().Value().(float64)
		case "RecvErrPps":
			linkResources[linkID].RecvErrPps += tableResult.Record().Value().(float64)
		case "RecvDropPps":
			linkResources[linkID].RecvDropPps += tableResult.Record().Value().(float64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return linkResources, nil
}

func GetPeriodLinkResourceDatas(periodExpr string, linkID string) ([]*model.LinkResource, error) {
	var linkResources []*model.LinkResource
	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
			from(bucket:"%s") |>
			range(start: -%s) |>
			filter(fn: (r) => r._measurement == "link_performance" and r.link_id == "%s") 
			`, config.GlobalConfig.Dependency.InfluxdbBucket, periodExpr, linkID,
		),
	)
	if err != nil {
		logrus.Errorf("Get Period Link Resource Data Error: %s", err.Error())
		return linkResources, err
	}
	for tableResult.Next() {
		var target *model.LinkResource
		for i := len(linkResources) - 1; i >= 0; i-- {
			if linkResources[i].Time == tableResult.Record().Time() {
				target = linkResources[i]
				break
			}
		}
		if target == nil {
			target = &model.LinkResource{}
			linkResources = append(linkResources, target)
		}
		switch tableResult.Record().Field() {
		case "SendBps":
			target.SendBps += tableResult.Record().Value().(float64)
			target.Time = tableResult.Record().Time()
		case "RecvBps":
			target.RecvBps += tableResult.Record().Value().(float64)
		case "SendPps":
			target.SendPps += tableResult.Record().Value().(float64)
		case "RecvPps":
			target.RecvPps += tableResult.Record().Value().(float64)
		case "SendErrPps":
			target.SendErrPps += tableResult.Record().Value().(float64)
		case "SendDropPps":
			target.SendDropPps += tableResult.Record().Value().(float64)
		case "RecvErrPps":
			target.RecvErrPps += tableResult.Record().Value().(float64)
		case "RecvDropPps":
			target.RecvDropPps += tableResult.Record().Value().(float64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return linkResources, nil
}

func GetPeriodAllLinkResourceDatas(periodStr string) (map[string][]*model.LinkResource, error) {
	linkResources := map[string][]*model.LinkResource{}

	tableResult, err := utils.InfluxDBQueryAPI.Query(
		context.Background(),
		fmt.Sprintf(
			`
				from(bucket:"%s") |>
				range(start: -%s) |>
				filter(fn: (r) => r._measurement == "link_performance") 
				`, config.GlobalConfig.Dependency.InfluxdbBucket, periodStr,
		),
	)
	if err != nil {
		logrus.Errorf("Get Period All Link Metrics Data Error: %s", err.Error())
		return linkResources, err
	}
	for tableResult.Next() {
		linkID := tableResult.Record().ValueByKey("link_id").(string)
		var target *model.LinkResource
		for i := len(linkResources[linkID]) - 1; i >= 0; i-- {
			if linkResources[linkID][i].Time == tableResult.Record().Time() {
				target = linkResources[linkID][i]
				break
			}
		}
		if target == nil {
			target = &model.LinkResource{}
			linkResources[linkID] = append(linkResources[linkID], target)
		}
		switch tableResult.Record().Field() {
		case "SendBps":
			target.SendBps += tableResult.Record().Value().(float64)
			target.Time = tableResult.Record().Time()
		case "RecvBps":
			target.RecvBps += tableResult.Record().Value().(float64)
		case "SendPps":
			target.SendPps += tableResult.Record().Value().(float64)
		case "RecvPps":
			target.RecvPps += tableResult.Record().Value().(float64)
		case "SendErrPps":
			target.SendErrPps += tableResult.Record().Value().(float64)
		case "SendDropPps":
			target.SendDropPps += tableResult.Record().Value().(float64)
		case "RecvErrPps":
			target.RecvErrPps += tableResult.Record().Value().(float64)
		case "RecvDropPps":
			target.RecvDropPps += tableResult.Record().Value().(float64)
		default:
			logrus.Warnf("Unknow InfluxDB Field: %s", tableResult.Record().Field())
		}
	}
	return linkResources, nil
}
