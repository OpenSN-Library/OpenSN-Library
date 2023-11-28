package config

import (
	"os"
	"strconv"
)

var (
	EtcdAddr       = ""
	EtcdPort       = 2379
	RedisAddr      = ""
	RedisPort      = 6379
	RedisDBIndex   = 0
	RedisPassword  = ""
	DockerSockPath = ""
	NodeIndex      = 0
	RedisImageName = ""
	EtcdImageName = ""
)

func init() {
	EtcdPort, _ = strconv.Atoi(os.Getenv("ETCD_PORT"))
	RedisPort, _ = strconv.Atoi(os.Getenv("REDIS_PORT"))
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	RedisDBIndex, _ = strconv.Atoi(os.Getenv("REDIS_DB_INDEX"))
}
