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

func GetLinkInfo(index int, linkID string) (*model.LinkBase, error) {
	linkInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, index)
	linkInfoKey := fmt.Sprintf("%s/%s", linkInfoKeyBase, linkID)
	etcdLinkInfo, err := utils.EtcdClient.Get(
		context.Background(),
		linkInfoKey,
	)

	if err != nil {
		return nil, fmt.Errorf("get link %s info from etcd error: %s", linkID, err.Error())
	}

	if len(etcdLinkInfo.Kvs) <= 0 {
		return nil, fmt.Errorf("link info of %s not found", linkID)
	}

	v := new(model.LinkBase)
	err = json.Unmarshal(etcdLinkInfo.Kvs[0].Value, v)
	if err != nil {
		err := fmt.Errorf(
			"unable to parse link info from etcd value %s, Error:%s",
			string(etcdLinkInfo.Kvs[0].Value),
			err.Error(),
		)
		return nil, err
	}
	return v, nil
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

func AddLinkInfo(nodeIndex int, linkInfo *model.LinkBase) error {
	linkInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, nodeIndex)
	linkInfoKey := fmt.Sprintf("%s/%s", linkInfoKeyBase, linkInfo.GetLinkID())
	linkSeq, err := json.Marshal(linkInfo)
	if err != nil {
		err = fmt.Errorf("format link info of %s error: %s", linkInfo.GetLinkID(), err.Error())
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		linkInfoKey,
		string(linkSeq),
	)

	if err != nil {
		err := fmt.Errorf("add link %s info to etcd error:%s", linkInfo.GetLinkID(), err.Error())
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

func UpdateLinkInfo(nodeIndex int, linkID string, update func(*model.LinkBase) error) error {
	linkInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, nodeIndex)
	linkInfoKey := fmt.Sprintf("%s/%s", linkInfoKeyBase, linkID)
	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		etcdOldLink := s.Get(linkInfoKey)
		updateLink := new(model.LinkBase)
		json.Unmarshal([]byte(etcdOldLink), updateLink)
		err := update(updateLink)
		if err != nil {
			return fmt.Errorf("update local new link of %s error: %s", linkInfoKey, err.Error())
		}
		etcdNewLink, err := json.Marshal(updateLink)
		if err != nil {
			return fmt.Errorf("format local new link of %s error: %s", linkInfoKey, err.Error())
		}
		s.Put(linkInfoKey, string(etcdNewLink))
		return nil
	})
	return err
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
