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
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

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

func GetLinkList(nodeIndex int) ([]model.Link, error) {
	var linkList []model.Link
	instanceInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, nodeIndex)
	nodeListEtcd, err := utils.EtcdClient.Get(
		context.Background(),
		instanceInfoKeyBase,
		clientv3.WithPrefix(),
	)
	if err != nil {
		err := fmt.Errorf("get link list from etcd error:%s", err.Error())
		return nil, err
	}

	for _, v := range nodeListEtcd.Kvs {
		linkInfo, err := link.ParseLinkFromBytes(v.Value)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to parse link info from etcd value %s, Error:%s", string(v.Value), err.Error())
			logrus.Errorf(errMsg)
			continue
		}
		linkList = append(linkList, linkInfo)
	}
	return linkList, nil
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

func UpdateLinkInfoIfExist(nodeIndex int, linkID string, update func(*model.LinkBase) error) error {
	linkInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, nodeIndex)
	linkInfoKey := fmt.Sprintf("%s/%s", linkInfoKeyBase, linkID)
	_, err := concurrency.NewSTM(utils.EtcdClient, func(s concurrency.STM) error {
		etcdOldLink := s.Get(linkInfoKey)
		updateLink := new(model.LinkBase)
		json.Unmarshal([]byte(etcdOldLink), updateLink)
		if updateLink.GetLinkID() == "" {
			return nil
		}

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

func RemoveLink(nodeIndex int, linkID string) error {
	linkInfoKeyBase := fmt.Sprintf(key.NodeLinkListKeyTemplate, nodeIndex)
	linkInfoKey := fmt.Sprintf("%s/%s", linkInfoKeyBase, linkID)
	_, err := utils.EtcdClient.Delete(context.Background(), linkInfoKey)
	if err != nil {
		return fmt.Errorf("remove link of %s error: %s", linkInfoKey, err.Error())
	}
	return nil
}

func GetLinkListParameters(nodeIndex int) (map[string]map[string]int64, error) {
	parameterMap := map[string]map[string]int64{}
	linkParameterKeyBase := fmt.Sprintf(key.NodeLinkParameterKeyTemplate, nodeIndex)
	nodeListEtcd, err := utils.EtcdClient.Get(
		context.Background(),
		linkParameterKeyBase,
		clientv3.WithPrefix(),
	)
	if err != nil {
		err := fmt.Errorf("get link parameter list from etcd error:%s", err.Error())
		return nil, err
	}

	for _, v := range nodeListEtcd.Kvs {
		parameterInfo := make(map[string]int64)
		err = json.Unmarshal(v.Value, &parameterInfo)
		if err != nil {
			errMsg := fmt.Sprintf("Unable to parse link parameter info from etcd value %s, Error:%s", string(v.Value), err.Error())
			logrus.Error(errMsg)
			continue
		}
		linkID, _ := utils.GetEtcdLastKey(string(v.Key))
		parameterMap[linkID] = parameterInfo
	}
	return parameterMap, nil
}

func GetLinkParameter(nodeIndex int, linkID string) (map[string]int64, error) {
	var parameter map[string]int64
	linkParameterKeyBase := fmt.Sprintf(key.NodeLinkParameterKeyTemplate, nodeIndex)
	linkParameterKey := fmt.Sprintf("%s/%s", linkParameterKeyBase, linkID)
	etcdParameterInfo, err := utils.EtcdClient.Get(
		context.Background(),
		linkParameterKey,
	)

	if err != nil {
		return nil, fmt.Errorf("get link %s info from etcd error: %s", linkID, err.Error())
	}

	if len(etcdParameterInfo.Kvs) <= 0 {
		return nil, fmt.Errorf("link info of %s not found", linkID)
	}
	err = json.Unmarshal(etcdParameterInfo.Kvs[0].Value, &parameter)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to parse link parameter info from etcd value %s, Error:%s", string(etcdParameterInfo.Kvs[0].Value), err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	return parameter, nil
}

func UpdateLinkParameter(nodeIndex int, linkID string, parameter map[string]int64) error {
	linkParameterKeyBase := fmt.Sprintf(key.NodeLinkParameterKeyTemplate, nodeIndex)
	linkParameterKey := fmt.Sprintf("%s/%s", linkParameterKeyBase, linkID)
	parameterSeq, err := json.Marshal(parameter)
	if err != nil {
		err = fmt.Errorf("format link parameter of %s error: %s", linkID, err.Error())
		return err
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		linkParameterKey,
		string(parameterSeq),
	)

	if err != nil {
		err := fmt.Errorf("add link %s parameter to etcd error:%s", linkID, err.Error())
		return err
	}
	return nil
}
