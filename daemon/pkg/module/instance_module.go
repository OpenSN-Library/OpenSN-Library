package module

import (
	"NodeDaemon/data"
	"NodeDaemon/model"
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
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var StopTimeoutSecond = 20

func CreateContainer(instance *model.Instance) error {
	containerID := fmt.Sprintf("%s_%s", instance.Type, instance.InstanceID)
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

		_, err := utils.DockerClient.ContainerCreate(
			context.Background(),
			containerConfig,
			hostConfig,
			nil,
			nil,
			containerID,
		)

		if err != nil {
			return err
		}

		return err
	}, 4)

	if err != nil {
		return fmt.Errorf("create container %s Error: %s", containerID, err.Error())
	}
	return nil
}

func StartContainer(instance *model.Instance) (int, error) {

	containerID := fmt.Sprintf("%s_%s", instance.Type, instance.InstanceID)
	pid := 0
	err := utils.DoWithRetry(func() error {
		return utils.DockerClient.ContainerStart(
			context.Background(),
			containerID,
			types.ContainerStartOptions{},
		)
	}, 4)
	if err != nil {
		return pid, fmt.Errorf("start container %s error: %s", containerID, err.Error())
	}

	err = utils.DoWithRetry(func() error {
		info, err := utils.DockerClient.ContainerInspect(context.Background(), containerID)
		if err != nil {
			return err
		}
		pid = info.State.Pid
		return nil
	}, 4)

	if err != nil {
		return pid, fmt.Errorf("get pid of container %s error: %s", containerID, err.Error())
	}

	return pid, nil
}

func StopContainer(instance *model.Instance) error {

	containerID := fmt.Sprintf("%s_%s", instance.Type, instance.InstanceID)

	err := utils.DoWithRetry(func() error {
		return utils.DockerClient.ContainerStop(context.Background(), containerID, container.StopOptions{})
	}, 4)

	if err != nil {
		err := fmt.Errorf(
			"stop container %s error: %s",
			containerID,
			err.Error(),
		)
		return err
	}

	return nil
}

func RemoveContainer(instance *model.Instance) error {
	containerID := fmt.Sprintf("%s_%s", instance.Type, instance.InstanceID)
	err := utils.DoWithRetry(func() error {
		return utils.DockerClient.ContainerRemove(
			context.Background(),
			containerID,
			types.ContainerRemoveOptions{Force: true},
		)
	}, 4)

	if err != nil {
		err := fmt.Errorf(
			"stop container %s error: %s",
			containerID,
			err.Error(),
		)
		return err
	}

	return nil
}

func UpdateInstanceState(oldInstance, newInstance *model.Instance) error {
	var err error
	isRunning := oldInstance.IsRunning() || newInstance.IsRunning()
	isCreated := oldInstance.IsCreated() || newInstance.IsCreated()

	shouldRunning := newInstance.Start
	shouldCreate := newInstance.InstanceID != ""

	if isCreated != shouldCreate {
		if shouldCreate {
			err = CreateContainer(newInstance)
		} else {
			StopContainer(oldInstance)
			data.DeleteInstancePid(oldInstance.InstanceID)
			err = RemoveContainer(oldInstance)
		}
	}

	if err != nil {
		return fmt.Errorf("update create state error: %s", err.Error())
	}

	if shouldRunning != isRunning {
		if shouldRunning {
			var pid int
			pid, err = StartContainer(newInstance)
			data.SetInstancePid(newInstance.InstanceID, pid)
		} else if shouldCreate {
			err = StopContainer(oldInstance)
			data.DeleteInstancePid(oldInstance.InstanceID)
		}
	}
	if err != nil {
		return fmt.Errorf("update running state error: %s", err.Error())
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
						err := UpdateInstanceState(oldInstance, newInstance)
						if err != nil {
							errMsg := fmt.Sprintf("Update Instance %s State Error: %s", etcdKey, err.Error())
							logrus.Error(errMsg)
						} else {
							logrus.Infof("Update Instance %s State Success", etcdKey)
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
