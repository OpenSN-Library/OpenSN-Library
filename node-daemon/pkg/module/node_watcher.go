package module

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type NodeWatchModule struct {
	Base
}

func CreateNodeWatchTask() *NodeWatchModule {
	return &NodeWatchModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			wg:         new(sync.WaitGroup),
			daemonFunc: watchNodeChange,
			running:    false,
			ModuleName: "NodeWatcher",
		},
	}
}

func parseNodeChange(nodeList []int) error {
	var addIndexListStr = []string{}
	var keepSet = map[int]bool{}
	data.NodeMapLock.Lock()
	defer data.NodeMapLock.Unlock()

	for _, v := range nodeList {
		if _, ok := data.NodeMap[v]; ok {
			keepSet[v] = true
		} else {
			addIndexListStr = append(addIndexListStr, strconv.Itoa(v))
		}
	}
	var delIndexListStr []string
	var delIndexList []int
	for k := range data.NodeMap {
		if !keepSet[k] {
			delIndexListStr = append(delIndexListStr, strconv.Itoa(k))
			delIndexList = append(delIndexList, k)
		}
	}
	for _, v := range delIndexList {
		delete(data.NodeMap, v)
	}

	if len(delIndexListStr) > 0 {
		err := utils.DoWithRetry(func() error {
			deleteResp := utils.RedisClient.HDel(context.Background(), key.NodesKey, delIndexListStr...)
			if deleteResp.Err() != nil {
				errMsg := fmt.Sprintf("Delete Keys %v Error: %s", delIndexListStr, deleteResp.Err().Error())
				logrus.Error(errMsg)
				return errors.New(errMsg)
			}
			return nil
		}, 3)

		if err != nil {
			return err
		}
	}

	if len(addIndexListStr) > 0 {
		err := utils.DoWithRetry(func() error {
			getResp := utils.RedisClient.HMGet(context.Background(), key.NodesKey, addIndexListStr...)
			if getResp.Err() != nil {
				errMsg := fmt.Sprintf("Get New Node Infos %v Error: %s", delIndexListStr, getResp.Err().Error())
				logrus.Error(errMsg)
				return errors.New(errMsg)
			}

			for i, v := range getResp.Val() {
				if v != nil {
					infoStr := v.(string)
					newNodeInfo := new(model.Node)
					err := json.Unmarshal([]byte(infoStr), newNodeInfo)
					if err != nil {
						errMsg := fmt.Sprintf("Parse Node Info of %s to Struct Error : %s ", addIndexListStr[i], err.Error())
						logrus.Error(errMsg)
						continue
					}
					data.NodeMap[int(newNodeInfo.NodeID)] = newNodeInfo
				}
			}

			return nil
		}, 3)

		return err
	}
	return nil
}

func watchNodeChange(sigChan chan int, errChan chan error) {
	for {
		if utils.CheckEtcdServe() {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	res, err := utils.EtcdClient.Get(
		context.Background(),
		key.NodeIndexListKey,
	)

	if err != nil {
		errMsg := fmt.Sprintf("Init Node Index List Error: %s", err.Error())
		logrus.Error(errMsg)
	} else if len(res.Kvs) <= 0 {
		errMsg := "Init Node Index List Error: empty array"
		logrus.Error(errMsg)
	} else {
		var nodeList []int
		err := json.Unmarshal(res.Kvs[0].Value, &nodeList)
		if err != nil {
			errMsg := fmt.Sprintf("Parse Node Index List Error: %s", res.Kvs[0].Value)
			logrus.Error(errMsg)
		} else {
			err = parseNodeChange(nodeList)
			if err != nil {
				errMsg := fmt.Sprintf("Parse Node Change Error: %s", err.Error())
				logrus.Error(errMsg)
			}
		}

	}

	for {
		watchChann := utils.EtcdClient.Watch(
			context.Background(),
			key.NodeIndexListKey,
		)

		select {
		case res := <-watchChann:
			if len(res.Events) <= 0 {
				continue
			}
			nodeListStr := res.Events[0].Kv.Value
			var nodeList []int
			err := json.Unmarshal([]byte(nodeListStr), &nodeList)
			if err != nil {
				errMsg := fmt.Sprintf("Parse Node Index List Error: %s", nodeListStr)
				logrus.Error(errMsg)
				continue
			}
			err = parseNodeChange(nodeList)
			if err != nil {
				errMsg := fmt.Sprintf("Parse Node Change Error: %s", err.Error())
				logrus.Error(errMsg)
			}
		case sig := <-sigChan:
			if sig == signal.STOP_SIGNAL {
				return
			}
		}
	}
}
