package utils_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedis(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", "10.134.148.56", 6379),
		Password: "123456", // no password set
		DB:       0,  // use default DB
	})
	cmdResp := rdb.HMSet(
		context.Background(),
		"bbb",
		map[string]string {
			"a":"b",
			"c":"d",
			"e":"f",
		},
	)

	if cmdResp.Err() != nil {
		t.Error(cmdResp.Err())
	}

	getResp := rdb.HMGet(context.Background(),"bbb")

	if getResp.Err() != nil {
		t.Error(getResp.Err())
	}
	
	fmt.Println(getResp.Val())

}