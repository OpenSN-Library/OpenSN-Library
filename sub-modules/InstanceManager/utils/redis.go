package utils

import (
	"InstanceManager/config"
	"fmt"
	redis "github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisAddr, config.RedisPort),
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDBIndex,  // use default DB
	})
	RedisClient = rdb
}
