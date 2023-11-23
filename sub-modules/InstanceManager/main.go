package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	fmt.Println("Start")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"10.134.148.56:2379",
		},
		DialTimeout: time.Second,
	})

	if err != nil {
		panic(err)
	}
	fmt.Println("Client Success")
	resp, err := cli.Get(context.Background(), "ab")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
