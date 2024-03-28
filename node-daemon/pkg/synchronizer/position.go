package synchronizer

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetAllInstancePosition() (map[string]model.Position, error) {
	positionBaseKey := key.InstancePositionKey
	positionList, err := utils.EtcdClient.Get(
		context.Background(),
		positionBaseKey,
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, fmt.Errorf("get all instance position error: %s", err.Error())
	}
	positionMap := make(map[string]model.Position)
	for _, v := range positionList.Kvs {
		position := model.Position{}
		err = json.Unmarshal(v.Value, &position)
		if err != nil {
			return nil, fmt.Errorf("unmarshal position error: %s", err.Error())
		}
		instanceID, _ := utils.GetEtcdLastKey(string(v.Key))
		positionMap[instanceID] = position
	}
	return positionMap, nil
}

func GetInstancePosition(instanceID string) (model.Position, error) {
	positionBaseKey := key.InstancePositionKey
	positionKey := fmt.Sprintf("%s/%s", positionBaseKey, instanceID)
	positionInfo, err := utils.EtcdClient.Get(
		context.Background(),
		positionKey,
	)

	if err != nil {
		return model.Position{}, fmt.Errorf("get instance %s info error: %s", instanceID, err.Error())
	}

	if len(positionInfo.Kvs) <= 0 {
		return model.Position{}, fmt.Errorf("instance %s position not found", instanceID)
	}

	position := model.Position{}
	err = json.Unmarshal(positionInfo.Kvs[0].Value, &position)
	if err != nil {
		return model.Position{}, fmt.Errorf("unmarshal position error: %s", err.Error())
	}
	return position, nil
}
