package utils

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/namespaces"
	"github.com/sirupsen/logrus"
)

var ContainerdClient *containerd.Client
var ContainerdNamespaceContext context.Context

func InitContainerdClient() error {
	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		logrus.Errorf("Init Containerd Client Error :%s", err.Error())
		return err
	}
	ContainerdClient = client
	ContainerdNamespaceContext = namespaces.WithNamespace(context.Background(), "emu")
	return nil
}
