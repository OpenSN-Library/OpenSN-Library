package utils

import (
	"NodeDaemon/config"
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
const DefaultLockTime = 2*time.Second
const DefaultLockGap = 500*time.Millisecond

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisAddr, config.RedisPort),
		Password: config.RedisPassword, // no password set
		DB:       config.RedisDBIndex,  // use default DB
	})
	RedisClient = rdb
}

func CheckRedisServe() bool {
	if RedisClient == nil {
			return false
	}
	resp := RedisClient.ClientList(context.Background())
	if resp.Err() != nil {
			return false
	}
	return true
}

func LockKeyWithTimeout(key string,timeout time.Duration) bool {
	lockKey := fmt.Sprintf("Lock_%s",key)
	successChann := make(chan bool)
	go func () {
		redisResp := RedisClient.SetNX(context.Background(),lockKey,"1",DefaultLockTime)
		if redisResp.Err() == nil {
			successChann <- redisResp.Val()
		}
		time.Sleep(DefaultLockGap)
	}()

	select {
	case success := <- successChann:
		return success
	case <- time.After(timeout):
		return false
	}
}