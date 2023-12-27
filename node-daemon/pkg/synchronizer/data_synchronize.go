package synchronizer

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/link"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

func GetNodeInstanceList(index int) ([]string, error) {
	var list []string
	etcdResp, err := utils.EtcdClient.Get(
		context.Background(),
		fmt.Sprintf(key.NodeInstanceListKeyTemplate, index),
	)
	logrus.Infof("Get Instance List of node %d Success.", index)
	if err != nil {
		errMsg := fmt.Sprintf("Get Instance List of node %d error: %s", index, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	err = json.Unmarshal(etcdResp.Kvs[0].Value, &list)
	if err != nil {
		errMsg := fmt.Sprintf("Parse Instance List of node %d error: %s", index, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	return list, err
}

func PostNodeInstanceList(index int, list []string) error {
	newListStr, err := json.Marshal(list)
	if err != nil {
		errMsg := fmt.Sprintf("Serialize Instance List of Node %d Error: %s", index, err.Error())
		logrus.Error(errMsg)
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		fmt.Sprintf(key.NodeInstanceListKeyTemplate, index),
		string(newListStr),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Update Instance List of Node %d to Etcd Error: %s", index, err.Error())
		logrus.Error(errMsg)
		return err
	}
	return nil
}

func AddInstanceInfosToNode(index int, instanceInfos []*model.InstanceConfig, namespace string, originInstanceList []string) ([]string, error) {
	var idSet = map[string]bool{}
	var redisValueArray []string

	for _, v := range originInstanceList {
		idSet[v] = true
	}

	for _, config := range instanceInfos {

		if idSet[config.InstanceID] {
			logrus.Warnf("Instance %s is already in Node %d.", config.InstanceID, index)
			continue
		}

		logrus.Infof("Set Instance %s to node %d Success.", config.InstanceID, index)
		info := model.Instance{
			Config:    *config,
			NodeID:    uint32(index),
			Namespace: namespace,
		}
		infoBytes, err := json.Marshal(info)
		if err != nil {
			errMsg := fmt.Sprintf("Serialize Instance Info of %s error: %s", info.Config.InstanceID, err.Error())
			logrus.Error(errMsg)
			return nil, err
		}
		redisValueArray = append(redisValueArray, config.InstanceID, string(infoBytes))
		originInstanceList = append(originInstanceList, config.InstanceID)
	}
	if len(redisValueArray) > 0 {
		setResp := utils.RedisClient.HMSet(
			context.Background(),
			fmt.Sprintf(key.NodeInstancesKeyTemplate, index),
			redisValueArray,
		)
		if setResp.Err() != nil {
			errMsg := fmt.Sprintf("Update Instances Info %d to Redis error: %s", index, setResp.Err().Error())
			logrus.Error(errMsg)
			return nil, setResp.Err()
		}
	}
	return originInstanceList, nil
}

func DelInstanceInfosFromNode(index int, instanceIDs []string, namespace string, originInstanceList []string) ([]string, error) {
	var idSet = map[string]bool{}
	var delSet = map[string]bool{}
	for _, v := range originInstanceList {
		idSet[v] = true
	}

	for _, id := range instanceIDs {
		if !idSet[id] {
			logrus.Warnf("Instance %s is not in Node %d.", id, index)
			continue
		}
		delSet[id] = true
	}

	delList := make([]string, 0, len(delSet))
	keepList := make([]string, 0, len(originInstanceList)-len(delSet))
	for _, v := range originInstanceList {
		if delSet[v] {
			delList = append(delList, v)
		} else {
			keepList = append(keepList, v)
		}
	}

	if len(delSet) > 0 {
		setResp := utils.RedisClient.HDel(
			context.Background(),
			fmt.Sprintf(key.NodeInstancesKeyTemplate, index),
			delList...,
		)
		if setResp.Err() != nil {
			errMsg := fmt.Sprintf("Update Instances Info %d to Redis error: %s", index, setResp.Err().Error())
			logrus.Error(errMsg)
			return nil, setResp.Err()
		}
	}
	return keepList, nil
}

func GetNodeLinkList(index int) ([]string, error) {
	var list []string
	etcdResp, err := utils.EtcdClient.Get(
		context.Background(),
		fmt.Sprintf(key.NodeLinkListKeyTemplate, index),
	)
	logrus.Infof("Get Link List of node %d Success.", index)
	if err != nil {
		errMsg := fmt.Sprintf("Get Link List of node %d error: %s", index, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	err = json.Unmarshal(etcdResp.Kvs[0].Value, &list)
	if err != nil {
		errMsg := fmt.Sprintf("Parse Link List of node %d error: %s", index, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	return list, err
}

func PostNodeLinkList(index int, list []string) error {
	newListStr, err := json.Marshal(list)
	if err != nil {
		errMsg := fmt.Sprintf("Serialize Link List of Node %d Error: %s", index, err.Error())
		logrus.Error(errMsg)
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		fmt.Sprintf(key.NodeLinkListKeyTemplate, index),
		string(newListStr),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Update Link List of Node %d to Etcd Error: %s", index, err.Error())
		logrus.Error(errMsg)
		return err
	}
	return nil
}

func AddLinkInfosToNode(index int, linkInfos []*model.LinkConfig, namespace string, originLinkList []string) ([]string, error) {
	var idSet = map[string]bool{}
	var redisValueArray []string

	for _, v := range originLinkList {
		idSet[v] = true
	}

	for _, config := range linkInfos {

		if idSet[config.LinkID] {
			logrus.Warnf("Link %s is already in Node %d.", config.LinkID, index)
			continue
		}

		logrus.Infof("Set Instance %s to node %d Success.", config.LinkID, index)
		info, err := link.ParseLinkFromConfig(*config, index)
		if err != nil {
			errMsg := fmt.Sprintf("Create Link %s Type %s error: %s", config.LinkID, config.Type, err.Error())
			logrus.Error(errMsg)
			return nil, err
		}
		infoBytes, err := json.Marshal(info)
		if err != nil {
			errMsg := fmt.Sprintf("Serialize Instance Info of %s error: %s", config.LinkID, err.Error())
			logrus.Error(errMsg)
			return nil, err
		}
		redisValueArray = append(redisValueArray, config.LinkID, string(infoBytes))
		originLinkList = append(originLinkList, config.LinkID)
	}
	if len(redisValueArray) > 0 {
		setResp := utils.RedisClient.HMSet(
			context.Background(),
			fmt.Sprintf(key.NodeLinksKeyTemplate, index),
			redisValueArray,
		)
		if setResp.Err() != nil {
			errMsg := fmt.Sprintf("Update Instances Info %d to Redis error: %s", index, setResp.Err().Error())
			logrus.Error(errMsg)
			return nil, setResp.Err()
		}
	}
	return originLinkList, nil
}

func DelLinkInfosFromNode(index int, LinkIDs []string, namespace string, originLinkList []string) ([]string, error) {
	var idSet = map[string]bool{}
	var delSet = map[string]bool{}
	for _, v := range originLinkList {
		idSet[v] = true
	}

	for _, id := range LinkIDs {
		if !idSet[id] {
			logrus.Warnf("Link %s is not in Node %d.", id, index)
			continue
		}
		delSet[id] = true
	}

	delList := make([]string, 0, len(delSet))
	keepList := make([]string, 0, len(originLinkList)-len(delSet))
	for _, v := range originLinkList {
		if delSet[v] {
			delList = append(delList, v)
		} else {
			keepList = append(keepList, v)
		}
	}

	if len(delSet) > 0 {
		setResp := utils.RedisClient.HDel(
			context.Background(),
			fmt.Sprintf(key.NodeLinksKeyTemplate, index),
			delList...,
		)
		if setResp.Err() != nil {
			errMsg := fmt.Sprintf("Update Links Info %d to Redis error: %s", index, setResp.Err().Error())
			logrus.Error(errMsg)
			return nil, setResp.Err()
		}
	}
	return keepList, nil
}

func PostNamespaceInfo(ns *model.Namespace) error {
	nsBytes, err := json.Marshal(ns)

	if err != nil {
		errMsg := fmt.Sprintf("Serialize Namespace %s Infomation Error: %s", ns.Name, err.Error())
		logrus.Error(errMsg)
		return err
	}

	hsetResp := utils.RedisClient.HSet(
		context.Background(),
		key.NamespacesKey,
		ns.Name,
		string(nsBytes),
	)

	if hsetResp.Err() != nil {
		errMsg := fmt.Sprintf("Update Namespace %s Infomation Error: %s", ns.Name, hsetResp.Err().Error())
		logrus.Error(errMsg)
		return err
	}
	return nil
}
