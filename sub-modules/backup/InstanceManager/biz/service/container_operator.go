package service

import (
	"InstanceManager/data"
	"InstanceManager/model"
	"InstanceManager/utils"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
)

var StopTimeoutSecond = 3
func AddContainers(addList []string) error {
	for _, v := range addList {
		instance, ok := data.InstanceMap[v]
		if ok {
			utils.DoWithRetry(func() error {
				containerConfig := &container.Config{
					Hostname:    instance.Name,
					Image:       "ubuntu:22.04",
					Env:         []string{},
					StopTimeout: &StopTimeoutSecond,
					Cmd: []string{"tail","-f","/dev/null"},
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
					instance.Name,
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
						v.InstanceID,
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
						v.InstanceID,
						v.ContainerID,
						err.Error(),
					)
					logrus.Error(errMsg)
				}
				return err
			}, 2)
		} else {
			errMsg := fmt.Sprintf("Container id of Instance %s is Empty, Skipping...", v.InstanceID)
			logrus.Error(errMsg)
		}
	}
	return nil
}
