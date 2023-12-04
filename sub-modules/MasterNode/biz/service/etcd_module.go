package service

import (
	"MasterNode/config"
	"MasterNode/data"
	"MasterNode/utils"
	"context"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

const EtcdModuleContainerName = "emulator_etcd"

var EtcdInitVal = map[string]interface{}{}

type EtcdModule struct {
	ModuleBase
}

func etcdDaemonFunc(sigChann chan int, errChann chan error) {
	containerConfig := &container.Config{
		Image: config.EtcdImageName,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true,
	}

	containerInfo, err := utils.DockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, EtcdModuleContainerName)
	if err != nil {
		logrus.Error("Create etcd Container Error: ",err.Error())
		errChann <- err
		return
	}
	err = utils.DockerClient.ContainerStart(context.Background(), containerInfo.ID, types.ContainerStartOptions{})

	if err != nil {
		logrus.Error("Start etcd Container Error: ",err.Error())
		errChann <- err
		return
	}

	for {
		select {
		case sig := <-sigChann:
			if sig == data.STOP_SIGNAL {
				utils.DoWithRetry(func() error {
					return utils.DockerClient.ContainerStop(context.Background(), containerInfo.ID, container.StopOptions{})
				}, 3)
				return
			}
		case <-time.After(ModuleCheckGap):
			status, err := utils.DockerClient.ContainerInspect(context.Background(), containerInfo.ID)
			if err != nil {
				errChann <- err
				return
			}
			if !status.State.Running {
				logrus.Warn("Instance Daemon Dead, Restarting...")
				err = utils.DockerClient.ContainerStart(context.Background(), containerInfo.ID, types.ContainerStartOptions{})
				if err != nil {
					errChann <- err
					return
				}
			}
		}
	}
}

func CreateEtcdModuleTask() *RedisModule {
	return &RedisModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: etcdDaemonFunc,
			wg:         new(sync.WaitGroup),
		},
	}
}

