package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdClient *clientv3.Client

func InitEtcdClient(addr string, port int) error {
	cliConfig := clientv3.Config{
		Endpoints: []string{
			fmt.Sprintf("%s:%d", addr, port),
		},
		DialTimeout: time.Second,
	}
	cli, err := clientv3.New(cliConfig)
	if err != nil {
		logrus.Error("Init Etcd Client Err: ", err.Error())
		return err
	}

	EtcdClient = cli
	return nil
}

func CheckEtcdServe() bool {
	if EtcdClient == nil {
		return false
	}
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := EtcdClient.Status(timeoutCtx, EtcdClient.Endpoints()[0])
	return err == nil
}