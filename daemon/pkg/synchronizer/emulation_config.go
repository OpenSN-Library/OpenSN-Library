package synchronizer

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	
	"go.etcd.io/etcd/client/v3/concurrency"
)


func GetEmulationInfo() (*model.EmulationInfo, error) {
	etcdNodeInfo, err := utils.EtcdClient.Get(
		context.Background(),
		key.EmulationConfigKey,
	)

	if err != nil {
		err := fmt.Errorf("get emulation info error: %s", err.Error())
		return nil, err
	}

	if len(etcdNodeInfo.Kvs) <= 0 {
		newInfo, _ := json.Marshal(model.EmulationInfo{})

		utils.EtcdClient.Put(
			context.Background(),
			key.EmulationConfigKey,
			string(newInfo),
		)
		return &model.EmulationInfo{}, nil
	}

	v := new(model.EmulationInfo)
	err = json.Unmarshal(etcdNodeInfo.Kvs[0].Value, v)
	if err != nil {
		err := fmt.Errorf(
			"unable to parse emulation info from etcd value %s, error:%s",
			string(etcdNodeInfo.Kvs[0].Value),
			err.Error(),
		)
		return nil, err
	}
	return v, nil
}

func UpdateEmulationInfo(update func(*model.EmulationInfo) error) error {
	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		etcdEmulationInfo := s.Get(key.EmulationConfigKey)
		updateEmulationInfo := new(model.EmulationInfo)
		json.Unmarshal([]byte(etcdEmulationInfo), updateEmulationInfo)
		if updateEmulationInfo.TypeConfig == nil {
			updateEmulationInfo.TypeConfig = make(map[string]model.InstanceTypeConfig)
		}
		err := update(updateEmulationInfo)
		if err != nil {
			return fmt.Errorf("update local new emulation info error: %s", err.Error())
		}
		nodeSeq, err := json.Marshal(updateEmulationInfo)
		if err != nil {
			err = fmt.Errorf("format emulation config error: %s", err.Error())
			return err
		}
		s.Put(
			key.EmulationConfigKey,
			string(nodeSeq),
		)

		if err != nil {
			err := fmt.Errorf("update emulation info to etcd error: %s", err.Error())
			return err
		}
		return nil
	})

	return err
}
