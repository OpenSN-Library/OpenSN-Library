package utils

import (
	"NodeDaemon/config"

	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var DockerClient *dockerClient.Client

func init() {

	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithHost(config.DockerHost))
	
	if err != nil {
		logrus.Error("Init Docker Client Error:", err.Error())
		panic(err)
	}
	DockerClient = cli
}
