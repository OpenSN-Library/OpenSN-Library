package utils

import (
	"fmt"

	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var DockerClient *dockerClient.Client

func InitDockerClient(sockPath string) error{
	url := fmt.Sprintf("unix://%s",sockPath)
	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithHost(url))
	
	if err != nil {
		logrus.Error("Init Docker Client Error:", err.Error())
		return err
	}
	DockerClient = cli
	return nil
}
