package biz

import (
	"NodeDaemon/config"
	"NodeDaemon/utils"
	"NodeDaemon/utils/tools"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
)

const MasterNodeContainerName = "master_node"

type MasterNodeModule struct {
	ModuleBase
	containerID string
}

func (m *MasterNodeModule) IsSetupFinish() bool {
	url := fmt.Sprintf("http://%s:8080/platform/status", config.MasterAddress)
	_, err := http.Get(url)
	return err == nil
}

func masterDaemonFunc(sigChann chan int, errChann chan error) {
	containerConfig := &container.Config{
		Hostname: "master_node",
		Image:    config.MasterNodeImage,
		Env: []string{
			fmt.Sprintf("DOCKER_HOST=%s", config.DockerHost),
			fmt.Sprintf("REDIS_IMAGE=%s", config.RedisImage),
			fmt.Sprintf("ETCD_IMAGE=%s", config.EtcdImage),
			fmt.Sprintf("ETCD_PORT=%d", config.EtcdPort),
			fmt.Sprintf("REDIS_PORT=%d", config.RedisPort),
			fmt.Sprintf("REDIS_PASSWORD=%s", config.RedisPassword),
			fmt.Sprintf("REDIS_DB_INDEX=%d", config.RedisDBIndex),
			fmt.Sprintf("MASTER_ADDRESS=%s", config.MasterAddress),
		},
	}
	natPort, _ := nat.NewPort("tcp", "8080")
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			natPort: []nat.PortBinding{{HostPort: "8080"}},
		},
		AutoRemove: true,
		Privileged: true,
	}

	dockerPath, found := strings.CutPrefix(config.DockerHost, "unix://")
	if found {
		hostConfig.Binds = []string{
			fmt.Sprintf("%s:%s", dockerPath, dockerPath),
		}
	}

	containerInfo, err := utils.DockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, MasterNodeContainerName)
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

func CreateMasterNodeModuleTask() *MasterNodeModule {
	return &MasterNodeModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			runing:     false,
			daemonFunc: masterDaemonFunc,
			wg:         new(sync.WaitGroup),
		}, "",
	}
}
