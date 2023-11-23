package utils

import (
	"InstanceManager/config"
	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var DockerClient *dockerClient.Client

func init() {

	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithHost(config.DockerSockPath))

	if err != nil {
		logrus.Error("Init Docker Client Error:", err.Error())
		panic(err)
	}
	DockerClient = cli
}
