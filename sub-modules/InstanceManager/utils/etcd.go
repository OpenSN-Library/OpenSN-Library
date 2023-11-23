package utils

import (
	"InstanceManager/config"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var EtcdClient *clientv3.Client

func init() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			fmt.Sprintf("%s:%d", config.EtcdAddr, config.EtcdPort),
		},
		DialTimeout: time.Second,
	})
	if err != nil {
		logrus.Error("Init Etcd Client Err: ", err.Error())
		panic(err)
	}
	EtcdClient = cli
}
