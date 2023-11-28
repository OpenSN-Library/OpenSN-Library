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
)

func init() {
	EtcdAddr = os.Getenv("ETCD_ADDR")
	EtcdPort, _ = strconv.Atoi(os.Getenv("ETCD_PORT"))
	RedisAddr = os.Getenv("REDIS_ADDR")
	RedisPort, _ = strconv.Atoi(os.Getenv("REDIS_PORT"))
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	RedisDBIndex, _ = strconv.Atoi(os.Getenv("REDIS_DB_INDEX"))
	DockerSockPath = os.Getenv("DOCKER_SOCK_PATH")
}
