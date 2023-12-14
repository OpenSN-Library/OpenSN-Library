package module

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

func init() {
	getResp, err := utils.EtcdClient.Get(
		context.Background(),
		key.NodeInstanceListKeySelf,
	)
	if err != nil {
		errMsg := fmt.Sprintf("Check Node Instance List Initialized %s Error: %s", key.NodeInstancesKeySelf, err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}
	if len(getResp.Kvs) <= 0 {
		_, err := utils.EtcdClient.Put(
			context.Background(),
			key.NodeInstanceListKeySelf,
			"[]",
		)
		if err != nil {
			errMsg := fmt.Sprintf("Init Node Instance List %s Error: %s", key.NodeInstancesKeySelf, err.Error())
			logrus.Error(errMsg)
			panic(errMsg)
		}
	}
}

var StopTimeoutSecond = 3

func AddContainers(addList []string) error {
	for _, v := range addList {
		instance, ok := data.InstanceMap[v]
		if ok {
			utils.DoWithRetry(func() error {
				containerConfig := &container.Config{
					Hostname:    instance.Config.Name,
					Image:       "ubuntu:22.04",
					Env:         []string{},
					StopTimeout: &StopTimeoutSecond,
					Cmd:         []string{"tail", "-f", "/dev/null"},
				}
				hostConfig := &container.HostConfig{
					Privileged: true,
				}

				createResp, err := utils.DockerClient.ContainerCreate(
					context.Background(),
					containerConfig,
					hostConfig,
					nil,
					nil,
					instance.Config.Name,
				)
				if err != nil {
					logrus.Error("Create Container Error: ", err.Error())
				}
				data.InstanceMap[v].ContainerID = createResp.ID
				return err
			}, 2)
		}

		for _, v := range addList {
			utils.DoWithRetry(func() error {
				err := utils.DockerClient.ContainerStart(
					context.Background(),
					data.InstanceMap[v].ContainerID,
					types.ContainerStartOptions{},
				)
				if err != nil {
					logrus.Error("Start Container Error: ", err.Error())
				}
				return err
			}, 2)
		}
	}
	return nil
}

func DelContainers(delList []*model.Instance) error {
	for _, v := range delList {
		if v.ContainerID != "" {
			utils.DoWithRetry(func() error {
				err := utils.DockerClient.ContainerStop(context.Background(), v.ContainerID, container.StopOptions{})
				if err != nil {
					errMsg := fmt.Sprintf(
						"Stop Container of Instance %s Error, Container id is %s, err: %s",
						v.Config.InstanceID,
						v.ContainerID,
						err.Error(),
					)
					logrus.Error(errMsg)
				}
				return err
			}, 2)
			utils.DoWithRetry(func() error {
				err := utils.DockerClient.ContainerRemove(context.Background(), v.ContainerID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					errMsg := fmt.Sprintf(
						"Remove Container of Instance %s Error, Container id is %s, err: %s",
						v.Config.InstanceID,
						v.ContainerID,
						err.Error(),
					)
					logrus.Error(errMsg)
				}
				return err
			}, 2)
		} else {
			errMsg := fmt.Sprintf("Container id of Instance %s is Empty, Skipping...", v.Config.InstanceID)
			logrus.Error(errMsg)
		}
	}
	return nil
}

const InstanceModuleContainerName = "instance_manager"

type InstanceModule struct {
	Base
}

func parseResult(updateIDMap map[string]string) (addList []string, delList []*model.Instance, err error) {
	var delIDList []string
	for k := range updateIDMap {
		if _, ok := data.InstanceMap[k]; !ok {
			addList = append(addList, k)
		}
	}

	for k := range data.InstanceMap {
		if _, ok := updateIDMap[k]; !ok {
			delIDList = append(addList, k)
		}
	}

	for _, v := range delIDList {
		delList = append(delList, data.InstanceMap[v])
		delete(data.InstanceMap, v)
	}

	redisResponse := utils.RedisClient.HMGet(context.Background(), key.NodeInstanceListKeySelf, addList...)

	if redisResponse.Err() != nil {
		err = redisResponse.Err()
		logrus.Error("Get Instance Infos Error: ", err.Error())
		return
	}

	for i, v := range redisResponse.Val() {
		if v == nil {
			logrus.Error("Redis Result Empty, Redis Data May Crash, InstanceID:", addList[i])
			continue
		} else {
			addInstance := new(model.Instance)
			err := json.Unmarshal([]byte(v.(string)), addInstance)
			if err != nil {
				logrus.Error("Unmarshal Json Data Error, Redis Data May Crash: ", err.Error())
				continue
			}
			data.InstanceMap[addList[i]] = addInstance
		}
	}
	return
}

func watchInstanceDaemon(sigChan chan int, errChan chan error) {
watchLoop:
	for {
		watchChan := utils.EtcdClient.Watch(context.Background(), key.NodeNsKeySelf)
		select {
		case res := <-watchChan:
			if len(res.Events) < 1 {
				logrus.Error("Unexpected Node Instance Info List Length:", len(res.Events))
				continue
			}
			infoBytes := res.Events[0].Kv.Value
			updateIDMap := make(map[string]string)
			err := json.Unmarshal(infoBytes, &updateIDMap)
			if err != nil {
				logrus.Error("Parse Update Instance  String Info Error: ", err.Error())
			}
			addList, delList, err := parseResult(updateIDMap)
			if err != nil {
				logrus.Error("Parse Update Instance Info Error: ", err.Error())
			}
			err = DelContainers(delList)
			if err != nil {
				errMsg := fmt.Sprintf("Delete Containers %v Error: %s", delList, err.Error())
				logrus.Error(errMsg)
				errChan <- err

			}
			err = AddContainers(addList)
			if err != nil {
				errMsg := fmt.Sprintf("Add Containers %v Error: %s", delList, err.Error())
				logrus.Error(errMsg)
				errChan <- err
			}
		case <-sigChan:
			break watchLoop
		}
	}
}

func CreateInstanceModuleTask() *InstanceModule {
	return &InstanceModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: watchInstanceDaemon,
			wg:         new(sync.WaitGroup),
			ModuleName: "InstanceManage",
		},
	}
}
