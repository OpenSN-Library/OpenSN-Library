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

func GetInstanceList(nodeIndex int) ([]*model.Instance, error) {
	var instanceList []*model.Instance
	instanceInfoKeyBase := fmt.Sprintf(key.NodeInstanceListKeyTemplate, nodeIndex)
	nodeListEtcd, err := utils.EtcdClient.Get(
		context.Background(),
		instanceInfoKeyBase,
		clientv3.WithPrefix(),
	)
	if err != nil {
		err := fmt.Errorf("get instance list from etcd error:%s", err.Error())
		return nil, err
	}

	for _, v := range nodeListEtcd.Kvs {
		instanceInfo := new(model.Instance)
		err = json.Unmarshal(v.Value, instanceInfo)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to parse instance info from etcd value %s, Error:%s", string(v.Value), err.Error())
			logrus.Debug(errMsg)
			continue
		}
		instanceList = append(instanceList, instanceInfo)
	}
	return instanceList, nil
}


func AddInstanceInfo(nodeIndex int, instanceInfo *model.Instance) error {
	instanceInfoKeyBase := fmt.Sprintf(key.NodeInstanceListKeyTemplate, nodeIndex)
	instanceInfoKey := fmt.Sprintf("%s/%s", instanceInfoKeyBase, instanceInfo.InstanceID)
	instanceSeq, err := json.Marshal(instanceInfo)
	if err != nil {
		err = fmt.Errorf("format instance info of %s error: %s", instanceInfo.InstanceID, err.Error())
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		instanceInfoKey,
		string(instanceSeq),
	)

	if err != nil {
		err := fmt.Errorf("add instance %s info to etcd error:%s", instanceInfo.InstanceID, err.Error())
		return err
	}
	return nil
}

func UpdateInstanceInfo(nodeIndex int, instanceID string, update func(*model.Instance) error) error {
	instanceInfoKeyBase := fmt.Sprintf(key.NodeInstanceListKeyTemplate, nodeIndex)
	instanceInfoKey := fmt.Sprintf("%s/%s", instanceInfoKeyBase, instanceID)
	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		etcdOldInstance := s.Get(instanceInfoKey)
		instance := new(model.Instance)
		json.Unmarshal([]byte(etcdOldInstance), instance)
		err := update(instance)
		if err != nil {
			return fmt.Errorf("update local new instance of %s error: %s", instanceInfoKey, err.Error())
		}
		etcdNewInstance, err := json.Marshal(instance)
		if err != nil {
			return fmt.Errorf("format local new instance of %s error: %s", instanceInfoKey, err.Error())
		}
		s.Put(instanceInfoKey, string(etcdNewInstance))
		return nil
	})
	return err
}

func RemoveInstance(nodeIndex int, instanceID string) error {
	instanceInfoKeyBase := fmt.Sprintf(key.NodeInstanceListKeyTemplate, nodeIndex)
	instanceInfoKey := fmt.Sprintf("%s/%s", instanceInfoKeyBase, instanceID)
	_,err := utils.EtcdClient.Delete(context.Background(),instanceInfoKey)
	if err != nil {
		return fmt.Errorf("remove instance of %s error: %s", instanceInfoKey, err.Error())
	}
	return nil
}