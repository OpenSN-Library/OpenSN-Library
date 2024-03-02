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
	"go.etcd.io/etcd/client/v3/concurrency"
)

func GetInstanceRuntimeList(nodeIndex int) ([]*model.InstanceRuntime, error) {

	var instanceRuntimeList []*model.InstanceRuntime
	instanceRuntimeKeyBase := fmt.Sprintf(key.NodeInstanceRuntimeKeyTemplate, nodeIndex)
	nodeListEtcd, err := utils.EtcdClient.Get(
		context.Background(),
		instanceRuntimeKeyBase,
		clientv3.WithPrefix(),
	)
	if err != nil {
		err := fmt.Errorf("get instance list from etcd error:%s", err.Error())
		return nil, err
	}

	for _, v := range nodeListEtcd.Kvs {
		instanceRuntimeInfo := new(model.InstanceRuntime)
		err = json.Unmarshal(v.Value, instanceRuntimeInfo)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to parse instance info from etcd value %s, Error:%s", string(v.Value), err.Error())
			logrus.Debug(errMsg)
			continue
		}
		instanceRuntimeList = append(instanceRuntimeList, instanceRuntimeInfo)
	}
	return instanceRuntimeList, nil
}

func GetInstanceRuntime(index int, instanceID string) (*model.InstanceRuntime, error) {
	instanceRuntimeKeyBase := fmt.Sprintf(key.NodeInstanceRuntimeKeyTemplate, index)
	instanceRuntimeKey := fmt.Sprintf("%s/%s", instanceRuntimeKeyBase, instanceID)
	etcdInstanceRuntimeInfo, err := utils.EtcdClient.Get(
		context.Background(),
		instanceRuntimeKey,
	)

	if err != nil {
		return nil, fmt.Errorf("get instance %s runtime info from etcd error: %s", instanceID, err.Error())
	}

	if len(etcdInstanceRuntimeInfo.Kvs) <= 0 {
		return nil, fmt.Errorf("instance runtime info of %s Not Found", instanceID)
	}

	v := new(model.InstanceRuntime)
	err = json.Unmarshal(etcdInstanceRuntimeInfo.Kvs[0].Value, v)
	if err != nil {
		err := fmt.Errorf(
			"unable to parse instance runtime info from etcd value %s, Error:%s",
			string(etcdInstanceRuntimeInfo.Kvs[0].Value),
			err.Error(),
		)
		return nil, err
	}
	return v, nil
}

func UpdateInstanceRuntimeInfo(nodeIndex int, instanceID string, update func(*model.InstanceRuntime) error) error {
	instanceRuntimeKeyBase := fmt.Sprintf(key.NodeInstanceRuntimeKeyTemplate, nodeIndex)
	instanceRuntimeKey := fmt.Sprintf("%s/%s", instanceRuntimeKeyBase, instanceID)
	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		etcdOldInstanceRuntime := s.Get(instanceRuntimeKey)
		updateInstanceRuntime := new(model.InstanceRuntime)
		json.Unmarshal([]byte(etcdOldInstanceRuntime), updateInstanceRuntime)
		err := update(updateInstanceRuntime)
		if err != nil {
			return fmt.Errorf("update local new instance runtime of %s error: %s", instanceRuntimeKey, err.Error())
		}
		etcdNewInstanceRuntime, err := json.Marshal(updateInstanceRuntime)
		if err != nil {
			return fmt.Errorf("format local new instance runtime of %s error: %s", instanceRuntimeKey, err.Error())
		}
		s.Put(instanceRuntimeKey, string(etcdNewInstanceRuntime))
		return nil
	})
	return err
}

func RemoveInstanceRuntime(nodeIndex int, instanceID string) error {
	instanceRuntimeKeyBase := fmt.Sprintf(key.NodeInstanceRuntimeKeyTemplate, nodeIndex)
	instanceRuntimeKey := fmt.Sprintf("%s/%s", instanceRuntimeKeyBase, instanceID)
	_, err := utils.EtcdClient.Delete(context.Background(), instanceRuntimeKey)
	if err != nil {
		return fmt.Errorf("remove instance of %s error: %s", instanceRuntimeKey, err.Error())
	}
	return nil
}
