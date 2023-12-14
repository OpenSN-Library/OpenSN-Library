package config

import (
	"os"
	"strconv"
)

var (
	MasterAddress = ""
	EtcdAddr       = ""
	EtcdPort       = 2379
	RedisAddr      = ""
	RedisPort      = 6379
	RedisDBIndex   = 0
	RedisPassword  = "1145141919810"
	DockerHost = "unix:///var/run/docker.sock"
	RedisImageName = "satellite_emulator/redis"
	EtcdImageName = "satellite_emulator/etcd"
)

func GetConfigEnvNumber(name string,defaultVal int) int {
	env := os.Getenv(name)
	envNum,err := strconv.Atoi(env)
	if err != nil {
		return defaultVal
	}
	return envNum
}

func GetConfigEnvString(name string,defaultVal string) string {
	env := os.Getenv(name)
	if env == "" {
		return defaultVal
	}
	return env
}

func init() {
	EtcdPort = GetConfigEnvNumber("ETCD_PORT",EtcdPort)
	RedisPort = GetConfigEnvNumber("REDIS_PORT",RedisPort)
	RedisPassword = GetConfigEnvString("REDIS_PASSWORD",RedisPassword)
	RedisDBIndex = GetConfigEnvNumber("REDIS_DB_INDEX",RedisDBIndex)
	MasterAddress = GetConfigEnvString("MASTER_ADDRESS",MasterAddress)
	DockerHost = GetConfigEnvString("DOCKER_HOST",DockerHost)
	RedisImageName = GetConfigEnvString("REDIS_IMAGE",RedisImageName)
	EtcdImageName = GetConfigEnvString("ETCD_IMAGE",EtcdImageName)
	EtcdAddr = MasterAddress
	RedisAddr = MasterAddress
}
