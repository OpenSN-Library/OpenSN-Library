package utils

import (
	"InstanceManager/config"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdClient *clientv3.Client

func init() {
	cliConfig := clientv3.Config{
		Endpoints: []string{
			fmt.Sprintf("%s:%d", config.EtcdAddr, config.EtcdPort),
		},
		DialTimeout: time.Second,
	}
	cli, err := clientv3.New(cliConfig)
	if err != nil {
		logrus.Error("Init Etcd Client Err: ", err.Error())
		panic(err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = cli.Status(timeoutCtx, cliConfig.Endpoints[0])

	EtcdClient = cli
}
