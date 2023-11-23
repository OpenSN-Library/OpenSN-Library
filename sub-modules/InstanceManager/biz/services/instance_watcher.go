package services

import (
	"InstanceManager/config"
	"InstanceManager/utils"
	"context"
	"fmt"
)

func ParseResult() {

}

func WatchInstance() {
	watchChan := utils.EtcdClient.Watch(context.Background(), config.NodeNsKey)
	res := <-watchChan
	fmt.Println(res)
}
