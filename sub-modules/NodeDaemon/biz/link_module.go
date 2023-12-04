package biz

import (
	"NodeDaemon/config"
	"NodeDaemon/utils"
	"NodeDaemon/utils/tools"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

const LinkModuleContainerName = "link_manager"

type LinkModule struct {
	ModuleBase
}

func linkDaemonFunc(sigChann chan int, errChann chan error) {
	containerConfig := &container.Config{
		Hostname: "link_manager",
		Image:    config.LinkManagerImage,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true,
		Privileged:  true,
	}
	dockerPath, found := strings.CutPrefix(config.DockerHost, "unix://")
	if found {
		hostConfig.Binds = []string{
			fmt.Sprintf("%s:%s", dockerPath, dockerPath),
		}
	}
	containerInfo, err := utils.DockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, LinkModuleContainerName)
	if err != nil {
		errChann <- err
		return
	}
	err = utils.DockerClient.ContainerStart(context.Background(), containerInfo.ID, types.ContainerStartOptions{})

	if err != nil {
		errChann <- err
		return
	}

	for {
		select {
		case sig := <-sigChann:
			if sig == config.STOP_SIGNAL {
				tools.DoWithRetry(func() error {
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
				logrus.Warn("Link Daemon Dead, Restarting...")
				err = utils.DockerClient.ContainerStart(context.Background(), containerInfo.ID, types.ContainerStartOptions{})
				if err != nil {
					errChann <- err
					return
				}
			}
		}
	}
}

func CreateLinkModuleTask() *LinkModule {
	return &LinkModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: linkDaemonFunc,
			wg:         new(sync.WaitGroup),
		},
	}
}
