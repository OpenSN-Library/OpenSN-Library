package synchronizer

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func UpdateWebShellRequest(nodeIndex int, webShellRequest *model.WebShellAllocRequest) error {
	webShellRequestKeyBase := fmt.Sprintf(key.NodeWebshellRequestKeyTemplate, nodeIndex)
	webShellRequestKey := fmt.Sprintf("%s/%s", webShellRequestKeyBase, webShellRequest.WebShellID)
	webShellRequestSeq, err := json.Marshal(webShellRequest)
	if err != nil {
		return fmt.Errorf("update webshell request error: %s", err.Error())
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		webShellRequestKey,
		string(webShellRequestSeq),
	)
	if err != nil {
		return fmt.Errorf("update webshell request error: %s", err.Error())
	}
	return nil
}

func UpdateGetWebshellInfo(nodeIndex int, webShellID string, info *model.WebShellAllocInfo) error {
	info.WebShellID = webShellID
	webShellInfoKeyBase := fmt.Sprintf(key.NodeWebshellInfoKeyTemplate, nodeIndex)
	webShellInfoKey := fmt.Sprintf("%s/%s", webShellInfoKeyBase, webShellID)
	webShellInfoSeq, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("update webshell info error: %s", err.Error())
	}
	_, err = utils.EtcdClient.Put(
		context.Background(),
		webShellInfoKey,
		string(webShellInfoSeq),
	)
	if err != nil {
		return fmt.Errorf("update webshell info error: %s", err.Error())
	}
	return nil
}

func DelGetWebshellInfo(nodeIndex int, webShellID string) error {
	webShellInfoKeyBase := fmt.Sprintf(key.NodeWebshellInfoKeyTemplate, nodeIndex)
	webShellInfoKey := fmt.Sprintf("%s/%s", webShellInfoKeyBase, webShellID)
	_, err := utils.EtcdClient.Delete(context.Background(), webShellInfoKey)
	if err != nil {
		return fmt.Errorf("delete webshell info error: %s", err.Error())
	}
	return nil
}

func WaitWebshellInfo(nodeIndex int, webShellID string, timeout time.Duration) (*model.WebShellAllocInfo, error) {
	webShellInfoChan := make(chan *model.WebShellAllocInfo)
	webShellInfoKeyBase := fmt.Sprintf(key.NodeWebshellInfoKeyTemplate, nodeIndex)
	webShellInfoKey := fmt.Sprintf("%s/%s", webShellInfoKeyBase, webShellID)
	keepState := true
	go func() {
		for keepState {
			webShellInfoSeq, err := utils.EtcdClient.Get(context.Background(), webShellInfoKey)
			if err != nil {
				continue
			}
			if len(webShellInfoSeq.Kvs) == 0 {
				continue
			}
			info := new(model.WebShellAllocInfo)
			err = json.Unmarshal(webShellInfoSeq.Kvs[0].Value, info)
			if err != nil {
				continue
			}
			webShellInfoChan <- info
			return
		}
	}()
	select {
	case res := <-webShellInfoChan:
		return res, nil
	case <-time.After(timeout):
		keepState = false
		return nil, fmt.Errorf("wait webshell info timeout")
	}
}

func GetWebshellInfo(nodeIndex int, webShellID string) (*model.WebShellAllocInfo, error) {
	webShellInfoKeyBase := fmt.Sprintf(key.NodeWebshellInfoKeySelf, webShellID)
	webShellInfoKey := fmt.Sprintf("%s/%s", webShellInfoKeyBase, webShellID)
	webShellInfoSeq, err := utils.EtcdClient.Get(context.Background(), webShellInfoKey)
	if err != nil {
		return nil, fmt.Errorf("get webshell info error: %s", err.Error())
	}
	if len(webShellInfoSeq.Kvs) == 0 {
		return nil, fmt.Errorf("webshell info not found")
	}
	info := new(model.WebShellAllocInfo)
	err = json.Unmarshal(webShellInfoSeq.Kvs[0].Value, info)
	if err != nil {
		return nil, fmt.Errorf("parse webshell info error: %s", err.Error())
	}
	return info, nil
}
