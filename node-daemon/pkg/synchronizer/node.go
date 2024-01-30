package synchronizer

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetNodeList() ([]*model.Node, error) {

	var nodeList []*model.Node

	nodeListEtcd, err := utils.EtcdClient.Get(
		context.Background(),
		key.NodeIndexListKey,
		clientv3.WithPrefix(),
	)
	if err != nil {
		err := fmt.Errorf("get node list from etcd error:%s", err.Error())
		return nil, err
	}

	for _, v := range nodeListEtcd.Kvs {
		nodeInfo := new(model.Node)
		err = json.Unmarshal(v.Value, nodeInfo)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to parse node info from etcd value %s, Error:%s", string(v.Value), err.Error())
			logrus.Debug(errMsg)
			continue
		}
		nodeList = append(nodeList, nodeInfo)
	}
	return nodeList, nil
}

func AddNode(nodeInfo *model.Node) error {
	nodeInfoKey := fmt.Sprintf("%s/%d", key.NodeIndexListKey, nodeInfo.NodeIndex)
	nodeSeq, err := json.Marshal(nodeInfo)
	if err != nil {
		err = fmt.Errorf("format node info of %d error: %s", nodeInfo.NodeIndex, err.Error())
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		nodeInfoKey,
		string(nodeSeq),
	)

	if err != nil {
		err := fmt.Errorf("add node %d info to etcd error:%s", nodeInfo.NodeIndex, err.Error())
		return err
	}
	return nil
}

func DelNode(index int) error {
	nodeInfoKey := fmt.Sprintf("%s/%d", key.NodeIndexListKey, index)
	_, err := utils.EtcdClient.Delete(
		context.Background(),
		nodeInfoKey,
	)

	if err != nil {
		err := fmt.Errorf("delete node %d info from etcd error: %s", index, err.Error())
		return err
	}
	return nil
}

func GetNode(index int) (*model.Node, error) {
	nodeInfoKey := fmt.Sprintf("%s/%d", key.NodeIndexListKey, index)
	etcdNodeInfo, err := utils.EtcdClient.Get(
		context.Background(),
		nodeInfoKey,
	)

	if err != nil {
		err := fmt.Errorf("get node %d info from etcd error: %s", index, err.Error())
		return nil, err
	}

	if len(etcdNodeInfo.Kvs) <= 0 {
		return nil, fmt.Errorf("node %d not found", index)

	}

	v := new(model.Node)
	err = json.Unmarshal(etcdNodeInfo.Kvs[0].Value, v)
	if err != nil {
		err := fmt.Errorf(
			"unable to parse node info from etcd value %s, Error:%s",
			string(etcdNodeInfo.Kvs[0].Value),
			err.Error(),
		)
		return nil, err
	}
	return v, nil
}