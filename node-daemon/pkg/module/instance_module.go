package module

import (
	"NodeDaemon/model"
	"NodeDaemon/share/data"
	"NodeDaemon/share/dir"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

func InitInstanceData() {
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
			err := utils.DoWithRetry(func() error {
				containerConfig := &container.Config{
					Hostname:    instance.Config.InstanceID,
					Image:       instance.Config.Image,
					Env:         []string{},
					StopTimeout: &StopTimeoutSecond,
				}
				hostConfig := &container.HostConfig{
					Privileged:  true,
					NetworkMode: "none",
					Binds: []string{
						fmt.Sprintf("%s:%s", dir.MountShareData, "/share"),
					},
					Resources: container.Resources{
						NanoCPUs: 1e7,
						Memory:   6 * (1 << 20),
					},
				}

				createResp, err := utils.DockerClient.ContainerCreate(
					context.Background(),
					containerConfig,
					hostConfig,
					nil,
					nil,
					instance.Config.Name,
				)

				data.InstanceMap[v].ContainerID = createResp.ID
				
				return err
			}, 2)
			if err != nil {
				logrus.Errorf("Creates Container %s Error: %s", v, err.Error())
			} else {
				logrus.Infof("Create and Start Container %s of %s Success.", data.InstanceMap[v].ContainerID, v)
			}
		}

	}
	for _, v := range addList {
		_, ok := data.InstanceMap[v]
		if ok {
			err := utils.DoWithRetry(func() error {
				err := utils.DockerClient.ContainerStart(
					context.Background(),
					data.InstanceMap[v].ContainerID,
					types.ContainerStartOptions{},
				)
				if err != nil {
					return err
				}
				inspect, err := utils.DockerClient.ContainerInspect(context.Background(), data.InstanceMap[v].ContainerID)
				if err != nil {
					return err
				}
				data.InstanceMap[v].Pid = inspect.State.Pid
				return nil
			}, 2)
			if err != nil {
				logrus.Errorf("Start Container %s Error: %s", v, err.Error())
			}
		}
	}
	return nil
}

func DelContainers(delList []string) error {
	for _, instanceID := range delList {
		if instanceID != "" {
			v := data.InstanceMap[instanceID]
			utils.Spin(func() bool {
				res := true
				for _, v := range v.LinkIDs {
					info := data.LinkMap[v]

					if info != nil && info.IsEnabled() {
						logrus.Infof("Instance Check Link %s is %v", info.GetLinkID(), info)
						res = false
					}
				}
				return res
			}, 100*time.Millisecond)
			err := utils.DoWithRetry(func() error {
				return utils.DockerClient.ContainerStop(context.Background(), v.ContainerID, container.StopOptions{})
			}, 2)
			if err != nil {
				errMsg := fmt.Sprintf(
					"Stop Container of Instance %s Error, Container id is %s, err: %s",
					v.Config.InstanceID,
					v.ContainerID,
					err.Error(),
				)
				logrus.Error(errMsg)
			} else {
				sucMsg := fmt.Sprintf(
					"Stop Container of Instance %s Success, Container id is %s",
					v.Config.InstanceID,
					v.ContainerID,
				)
				logrus.Info(sucMsg)
			}
			utils.DoWithRetry(func() error {
				return utils.DockerClient.ContainerRemove(context.Background(), v.ContainerID, types.ContainerRemoveOptions{Force: true})
			}, 2)
			if err != nil {
				errMsg := fmt.Sprintf(
					"Remove Container of Instance %s Error, Container id is %s, err: %s",
					v.Config.InstanceID,
					v.ContainerID,
					err.Error(),
				)
				logrus.Error(errMsg)
			} else {
				sucMsg := fmt.Sprintf(
					"Remove Container of Instance %s Success, Container id is %s",
					v.Config.InstanceID,
					v.ContainerID,
				)
				logrus.Info(sucMsg)
			}
			delete(data.InstanceMap, instanceID)
		} else {
			errMsg := fmt.Sprintf("Container id of Instance %s is Empty, Skipping...", instanceID)
			logrus.Error(errMsg)
		}
	}
	return nil
}

const InstanceModuleContainerName = "instance_manager"

type InstanceModule struct {
	Base
}

func parseInstanceChange(updateIdList []string) (addList []string, delList []string, err error) {
	updateIDMap := make(map[string]bool)
	for _, v := range updateIdList {
		updateIDMap[v] = true
	}
	for k := range updateIDMap {
		if _, ok := data.InstanceMap[k]; !ok {
			addList = append(addList, k)
		}
	}

	for k := range data.InstanceMap {
		if ok := updateIDMap[k]; !ok {
			delList = append(delList, k)
		}
	}

	if len(addList) > 0 {

		redisResponse := utils.RedisClient.HMGet(context.Background(), key.NodeInstancesKeySelf, addList...)

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
	}
	return
}

func watchInstanceDaemon(sigChan chan int, errChan chan error) {
	InitInstanceData()
	for {
		ctx, cancel := context.WithCancel(context.Background())

		watchChan := utils.EtcdClient.Watch(ctx, key.NodeInstanceListKeySelf)

		for {
			select {
			case sig := <-sigChan:
				if sig == signal.STOP_SIGNAL {
					cancel()
					return
				}
			case res := <-watchChan:
				if len(res.Events) < 1 {
					logrus.Error("Unexpected Node Instance Info List Length:", len(res.Events))
					continue
				}
				infoBytes := res.Events[0].Kv.Value
				updateIDList := []string{}
				err := json.Unmarshal(infoBytes, &updateIDList)
				if err != nil {
					logrus.Error("Parse Update Instance  String Info Error: ", err.Error())
				}
				addList, delList, err := parseInstanceChange(updateIDList)
				if err != nil {
					logrus.Error("Parse Update Instance Info Error: ", err.Error())
				} else {
					logrus.Infof("Parse Update Instance Info Success: Addlist:%v,Dellist: %v", addList, delList)
				}
				go func() {
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
				}()
			}
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
