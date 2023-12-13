package data

import (
	"MasterNode/utils"
	"context"
	"time"
)

const ( // Etcd Keys
	NodeIndexListKey = "/node_index_list"
)

const (
	NodeInstanceListKey = "/node_%d/instance_list"
	NodeInstanceInfoKey = "node_%d_instances"
)

const ( // Redis Keys
	NodeHeartBeatKey = "node_heart_beat"
	NodesKey         = "nodes"
	NextNodeIndexKey = "next_node_index"
	NamespacesKey    = "namespaces"
)

func InitRedisData() error {
	for !utils.CheckRedisServe() {
		time.Sleep(500 * time.Millisecond)
	}
	msetResp := utils.RedisClient.HSet(context.Background(), NodeHeartBeatKey)

	if msetResp.Err() != nil {
		return msetResp.Err()
	}

	msetResp = utils.RedisClient.HSet(context.Background(), NodesKey)

	if msetResp.Err() != nil {
		return msetResp.Err()
	}

	setResp := utils.RedisClient.Set(context.Background(), NextNodeIndexKey, "1", 0)

	if setResp.Err() != nil {
		return setResp.Err()
	}

	msetResp = utils.RedisClient.HSet(context.Background(), NamespacesKey)

	if msetResp.Err() != nil {
		return msetResp.Err()
	}
	return nil
}

func InitEtcdData() error {
	for !utils.CheckEtcdServe() {
		time.Sleep(500 * time.Millisecond)
	}
	_, err := utils.EtcdClient.Put(context.Background(), NodeIndexListKey, "[]")
	if err != nil {
		return err
	}
	return nil
}
