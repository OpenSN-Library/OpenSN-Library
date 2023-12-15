package utils

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

const DefaultLockTime = 2 * time.Second
const DefaultLockGap = 500 * time.Millisecond

func InitRedisClient(addr,password string,port,index int) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", addr, port),
		Password: password, // no password set
		DB:       index,  // use default DB
	})
	RedisClient = rdb
}

func CheckRedisServe() bool {
	if RedisClient == nil {
		return false
	}
	resp := RedisClient.ClientList(context.Background())
	return resp.Err() == nil 
}

func LockKeyWithTimeout(key string, timeout time.Duration) bool {
	lockKey := fmt.Sprintf("Lock_%s", key)
	successChann := make(chan bool)
	go func() {
		redisResp := RedisClient.SetNX(context.Background(), lockKey, "1", DefaultLockTime)
		if redisResp.Err() == nil {
			successChann <- redisResp.Val()
		}
		time.Sleep(DefaultLockGap)
	}()

	select {
	case success := <-successChann:
		return success
	case <-time.After(timeout):
		return false
	}
}
