package services

import (
	"InstanceManager/data"
	"InstanceManager/model"
	"InstanceManager/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

func init() {
	getResp, err := utils.EtcdClient.Get(
		context.Background(),
		data.NodeInstanceListKey,
	)
	if err != nil {
		errMsg := fmt.Sprintf("Check Node Instance List Initialized%s Error: %s", data.NodeInstancesKey, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	if len(getResp.Kvs) <= 0 {
		_, err := utils.EtcdClient.Put(
			context.Background(),
			data.NodeInstanceListKey,
			"[]",
		)
		if err != nil {
			errMsg := fmt.Sprintf("Init Node Instance List %s Error: %s", data.NodeInstancesKey, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
	}
}

func ParseResult(updateIDMap map[string]string) (addList []string, delList []*model.Instance, err error) {
	var delIDList []string
	for k, _ := range updateIDMap {
		if _, ok := data.InstanceMap[k]; !ok {
			addList = append(addList, k)
		}
	}

	for k, _ := range data.InstanceMap {
		if _, ok := updateIDMap[k]; !ok {
			delIDList = append(addList, k)
		}
	}

	for _, v := range delIDList {
		delList = append(delList, data.InstanceMap[v])
		delete(data.InstanceMap, v)
	}

	redisResponse := utils.RedisClient.HMGet(context.Background(), data.NodeInstanceListKey, addList...)

	if redisResponse.Err() != nil {
		err = redisResponse.Err()
		logrus.Error("Get Instance Infos Error: ", err.Error())
		return
	}

	for i, v := range redisResponse.Val() {
		if v == nil {
			logrus.Error("Redis Result Empty, Redis Data May Crash, InstanceID:", addList[i])
			continue
		} else {
			addInstance := new(model.Instance)
			err := json.Unmarshal([]byte(v.(string)), addInstance)
			if err != nil {
				logrus.Error("Unmarshal Json Data Error, Redis Data May Crash: ", err.Error())
				continue
			}
			data.InstanceMap[addList[i]] = addInstance
		}
	}
	return
}

func WatchInstance(sigChan chan int) {
	for {
		watchChan := utils.EtcdClient.Watch(context.Background(), data.NodeNsKey)
		select {
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Instance Info List Length:", len(res.Events))
				continue
			}
			infoBytes := res.Events[0].Kv.Value
			updateIDMap := make(map[string]string)
			json.Unmarshal(infoBytes, updateIDMap)
			addList, delList, err := ParseResult(updateIDMap)
			if err != nil {
				logrus.Error("Parse Update Instance Info Error: ", err.Error())
			}
			DelContainers(delList)
			AddContainers(addList)
		case <-sigChan:
			break
		}
	}
}
