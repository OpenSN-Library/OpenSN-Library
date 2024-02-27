package module

import (
	"NodeDaemon/model"
	"NodeDaemon/pkg/synchronizer"
	"encoding/json"

	"NodeDaemon/share/dir"
	"NodeDaemon/share/key"
	"NodeDaemon/share/signal"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var StopTimeoutSecond = 3

func SetLinkEndPid(linkID string, instanceID string, pid int) error {

	err := synchronizer.UpdateLinkInfoIfExist(
		key.NodeIndex,
		linkID,
		func(lb *model.LinkBase) error {
			for endIndex, endInfo := range lb.EndInfos {
				if endInfo.EndNodeIndex == key.NodeIndex && endInfo.InstanceID == instanceID {
					lb.EndInfos[endIndex].InstancePid = pid
				}
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("update link end pid of %s in %d error: %s", linkID, key.NodeIndex, err.Error())
	}
	logrus.Debugf("Update Instance %s Pid of Link %s to %d", instanceID, linkID, pid)
	return nil
}

func StartContainer(instance *model.Instance) (*model.InstanceRuntime, error) {

	instanceRuntime := new(model.InstanceRuntime)
	instanceRuntime.InstanceID = instance.InstanceID
	err := utils.DoWithRetry(func() error {

		containerConfig := &container.Config{
			Hostname:    instance.InstanceID,
			Image:       instance.Image,
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
				NanoCPUs: instance.Resource.NanoCPU,
				Memory:   instance.Resource.MemoryByte,
			},
		}

		createResp, err := utils.DockerClient.ContainerCreate(
			context.Background(),
			containerConfig,
			hostConfig,
			nil,
			nil,
			instance.Name,
		)

		if err != nil {
			return err
		}

		instanceRuntime.ContainerID = createResp.ID
		return err
	}, 2)

	if err != nil {
		return nil, fmt.Errorf("create container of instance %s Error: %s", instance.InstanceID, err.Error())
	}

	err = utils.DoWithRetry(func() error {
		err := utils.DockerClient.ContainerStart(
			context.Background(),
			instanceRuntime.ContainerID,
			types.ContainerStartOptions{},
		)
		if err != nil {
			return err
		}
		inspect, err := utils.DockerClient.ContainerInspect(context.Background(), instanceRuntime.ContainerID)
		if err != nil {
			return err
		}
		instanceRuntime.Pid = inspect.State.Pid
		instanceRuntime.State = "Running"
		return nil
	}, 2)
	if err != nil {
		return nil, fmt.Errorf("start container %s error: %s", instanceRuntime.ContainerID, err.Error())
	}

	return instanceRuntime, err
}

func StopContainer(instance *model.Instance) (*model.InstanceRuntime, error) {

	instanceRuntime, err := synchronizer.GetInstanceRuntime(instance.NodeIndex, instance.InstanceID)

	if err != nil {
		return &model.InstanceRuntime{
			Pid:         0,
			State:       "Stop",
			ContainerID: "",
		}, fmt.Errorf("stop instance %s error: get instance runtime error: %s", instance.InstanceID, err.Error())
	}

	err = utils.DoWithRetry(func() error {
		return utils.DockerClient.ContainerStop(context.Background(), instanceRuntime.ContainerID, container.StopOptions{})
	}, 2)
	utils.DockerClient.NetworkConnect(context.Background(),"","",&network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			
		},
	})
	if err != nil {
		err := fmt.Errorf(
			"stop container %s, err: %s",
			instanceRuntime.ContainerID,
			err.Error(),
		)
		return nil, err
	}

	err = utils.DoWithRetry(func() error {
		return utils.DockerClient.ContainerRemove(
			context.Background(),
			instanceRuntime.ContainerID,
			types.ContainerRemoveOptions{Force: true},
		)
	}, 2)

	if err != nil {
		err := fmt.Errorf(
			"remove container %s of instance %s error: %s",
			instanceRuntime.ContainerID,
			instance.InstanceID,
			err.Error(),
		)
		return nil, err
	}

	return &model.InstanceRuntime{
		InstanceID:  instance.InstanceID,
		Pid:         0,
		State:       "Stop",
		ContainerID: "",
	}, err
}

func ParseLinkChange(oldInstance, newInstance *model.Instance, runtimeInfo *model.InstanceRuntime) error {
	var oldLinkMap map[string]model.ConnectionInfo
	var newLinkMap map[string]model.ConnectionInfo
	newLinkIDSet := map[string]bool{}

	if oldInstance.Start {
		oldLinkMap = oldInstance.Connections
	} else {
		oldLinkMap = make(map[string]model.ConnectionInfo)
	}

	if newInstance.Start {
		newLinkMap = newInstance.Connections
	} else {
		newLinkMap = make(map[string]model.ConnectionInfo)
	}

	for linkID := range newLinkMap {
		newLinkIDSet[linkID] = true
	}

	for linkID := range oldLinkMap {
		if !newLinkIDSet[linkID] {
			err := SetLinkEndPid(linkID, runtimeInfo.InstanceID, 0)
			if err != nil {
				return fmt.Errorf("set link %s end pid to 0 error: %s", linkID, err.Error())
			}
		} else {
			newLinkIDSet[linkID] = false
		}
	}
	for linkID, notExist := range newLinkIDSet {
		if notExist {
			err := SetLinkEndPid(linkID, runtimeInfo.InstanceID, runtimeInfo.Pid)
			if err != nil {
				return fmt.Errorf("set link %s end pid to 0 error: %s", linkID, err.Error())
			}
		}
	}
	return nil
}

const InstanceModuleContainerName = "instance_manager"

type InstanceModule struct {
	Base
}

func watchInstanceDaemon(sigChan chan int, errChan chan error) {
	for {
		ctx, cancel := context.WithCancel(context.Background())
		var keyLockMap sync.Map
		watchChan := utils.EtcdClient.Watch(ctx, key.NodeInstanceListKeySelf, clientv3.WithPrefix(), clientv3.WithPrevKV())

		for {
			select {
			case sig := <-sigChan:
				if sig == signal.STOP_SIGNAL {
					cancel()
					return
				}
			case res := <-watchChan:
				for _, v := range res.Events {
					go func(v *clientv3.Event) {
						etcdKey := ""
						oldInstance := new(model.Instance)
						newInstance := new(model.Instance)
						if v.PrevKv != nil && len(v.PrevKv.Value) > 0 {
							etcdKey = string(v.PrevKv.Key)
							err := json.Unmarshal(v.PrevKv.Value, oldInstance)
							if err != nil {
								errMsg := fmt.Sprintf("Parse Instance Info From %s Error: %s", etcdKey, err.Error())
								logrus.Error(errMsg)
								return
							}
						}
						if v.Kv != nil && len(v.Kv.Value) > 0 {
							etcdKey = string(v.Kv.Key)
							err := json.Unmarshal(v.Kv.Value, newInstance)
							if err != nil {
								errMsg := fmt.Sprintf("Parse Instance Info From %s Error: %s", etcdKey, err.Error())
								logrus.Error(errMsg)
								return
							}
						}
						lockAny, _ := keyLockMap.LoadOrStore(etcdKey, new(sync.Mutex))
						lock := lockAny.(*sync.Mutex)
						lock.Lock()
						defer lock.Unlock()
						var update *model.InstanceRuntime
						var err error
						if newInstance.Start != oldInstance.Start {
							if newInstance.Start {
								update, err = StartContainer(newInstance)
								if err != nil {
									errMsg := fmt.Sprintf("Start Container of %s Error: %s", etcdKey, err.Error())
									logrus.Error(errMsg)
									return
								}
								logrus.Infof("Start Container of %s Success.", etcdKey)
							} else {
								update, err = StopContainer(oldInstance)
								if err != nil {
									errMsg := fmt.Sprintf("Stop Container of %s Error: %s", etcdKey, err.Error())
									logrus.Error(errMsg)
									return
								}
								logrus.Infof("Stop Container of %s Success.", etcdKey)
							}
							err = synchronizer.UpdateInstanceRuntimeInfo(
								key.NodeIndex,
								update.InstanceID,
								func(ir *model.InstanceRuntime) error {
									*ir = *update
									return nil
								},
							)
							if err != nil {
								errMsg := fmt.Sprintf("Update Runtime info of %s Error: %s", etcdKey, err.Error())
								logrus.Error(errMsg)
								return
							}
						} else if newInstance.Start {
							update, err = synchronizer.GetInstanceRuntime(
								key.NodeIndex, newInstance.InstanceID,
							)
							if err != nil {
								errMsg := fmt.Sprintf("Get Instance Runtime Info of %s Error: %s", etcdKey, err.Error())
								logrus.Error(errMsg)
								return
							}
						}
						err = ParseLinkChange(oldInstance, newInstance, update)
						if err != nil {
							errMsg := fmt.Sprintf("Parse Link Change of %s Error: %s", update.InstanceID, err.Error())
							logrus.Error(errMsg)
						}
					}(v)
				}

			}
		}
	}
}

func CreateInstanceManagerModule() *InstanceModule {
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
