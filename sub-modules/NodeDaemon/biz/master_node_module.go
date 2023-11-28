package biz

import (
	"NodeDaemon/config"
	"NodeDaemon/utils"
	"context"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

type MasterNodeModule struct {
	ModuleBase
	containerID string
}

func (m *MasterNodeModule) IsSetupFinish() bool {
	return false
}

func masterDaemonFunc(sigChann chan int, errChann chan error) {
	containerConfig := &container.Config{

	}
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		AutoRemove: true,
	}
	
	containerInfo,err := utils.DockerClient.ContainerCreate(context.Background(),containerConfig,hostConfig,nil,nil,InstanceModuleContainerName)
	if err != nil {
		errChann <- err
		return
	}
	err = utils.DockerClient.ContainerStart(context.Background(),containerInfo.ID,types.ContainerStartOptions{})

	if err != nil {
		errChann <- err
		return
	}

	for {
		select {
		case sig := <- sigChann:
			if sig == config.STOP_SIGNAL {
				utils.DoWithRetry(func() error {
					return utils.DockerClient.ContainerStop(context.Background(),containerInfo.ID,container.StopOptions{})
				},3)
				return
			}
		case <- time.After(ModuleCheckGap):
			status,err := utils.DockerClient.ContainerInspect(context.Background(),containerInfo.ID)
			if err != nil {
				errChann <- err
				return
			}
			if !status.State.Running {
				logrus.Warn("Instance Daemon Dead, Restarting...")
				err = utils.DockerClient.ContainerStart(context.Background(),containerInfo.ID,types.ContainerStartOptions{})
				if err != nil {
					errChann <- err
					return
				}
			}
		}
	}
}

func CreateMasterNodeModuleTask() *MasterNodeModule {
	return &MasterNodeModule{
		ModuleBase{
			sigChan: make(chan int),
			errChan: make(chan error),
			runing: false,
			daemonFunc: masterDaemonFunc,
			wg: new(sync.WaitGroup),
		},"",
	}
}