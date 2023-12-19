package data

import (
	"NodeDaemon/config"
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"
)

func initNamespaceMap() error {
	getResp := utils.RedisClient.HGetAll(context.Background(), key.NamespacesKey)
	if getResp.Err() != nil {
		logrus.Errorf("Init Node Map Error: %s", getResp.Err().Error())
		return getResp.Err()
	}
	for name, nsInfo := range getResp.Val() {
		var nsObj = new(model.Namespace)
		err := json.Unmarshal([]byte(nsInfo), nsObj)

		if err != nil {
			logrus.Errorf("Parse Init Namesapce %s Info Error: %s",name, err.Error())
			continue
		}

		NamespaceMap[name] = nsObj

	}
	return nil

}

func initInstanceMap() error {
	getResp := utils.RedisClient.HGetAll(context.Background(), key.NodeInstancesKeySelf)
	if getResp.Err() != nil {
		logrus.Errorf("Init Instance Map Error: %s", getResp.Err().Error())
		return getResp.Err()
	}
	for indexStr, nodeInfo := range getResp.Val() {
		var nodeObj = new(model.Instance)
		err := json.Unmarshal([]byte(nodeInfo), nodeObj)

		if err != nil {
			logrus.Errorf("Parse Init Instance %s Info Error: %s",indexStr, err.Error())
			continue
		}

		InstanceMap[indexStr] = nodeObj

	}
	return nil
}

func initNodeMap() error {
	getResp := utils.RedisClient.HGetAll(context.Background(), key.NodesKey)
	if getResp.Err() != nil {
		logrus.Errorf("Init Node Map Error: %s", getResp.Err().Error())
		return getResp.Err()
	}
	for indexStr, nodeInfo := range getResp.Val() {
		var nodeObj = new(model.Node)
		err := json.Unmarshal([]byte(nodeInfo), nodeObj)

		if err != nil {
			logrus.Errorf("Parse Init Node %s Info Error: %s",indexStr, err.Error())
			continue
		}
		nodeIndex, err := strconv.Atoi(indexStr)

		if err != nil {
			logrus.Errorf("Parse Init Node Index %s Error: %s",indexStr, err.Error())
			continue
		}

		NodeMap[nodeIndex] = nodeObj

	}
	return nil
}

func InitData() error {
	if !config.GlobalConfig.App.IsServant {
		err := initNodeMap()

		if err != nil {
			return err
		}

		err = initNamespaceMap()

		if err != nil {
			return err
		}
	}
	return initInstanceMap()
}
