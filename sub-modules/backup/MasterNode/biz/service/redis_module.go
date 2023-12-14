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

const RedisModuleContainerName = "emulator_redis"

var RedisInitVal = map[string]interface{}{}

type RedisModule struct {
	ModuleBase
}

func redisDaemonFunc(sigChann chan int, errChann chan error) {
	containerConfig := &container.Config{
		Image: config.RedisImageName,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true,
	}

	containerInfo, err := utils.DockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, RedisModuleContainerName)
	if err != nil {
		logrus.Error("Create redis Container Error: ", err.Error())
		errChann <- err
		return
	}
	err = utils.DockerClient.ContainerStart(context.Background(), containerInfo.ID, types.ContainerStartOptions{})

	if err != nil {
		logrus.Error("Start redis Container Error: ", err.Error())
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

func CreateRedisModuleTask() *RedisModule {
	return &RedisModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: redisDaemonFunc,
			wg:         new(sync.WaitGroup),
		},
	}
}
