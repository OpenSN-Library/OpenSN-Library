package utils_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {
	cliConfig := clientv3.Config{
		Endpoints: []string{
			"10.134.148.56:2379",
		},
		DialTimeout: time.Second,
	}

	fmt.Println("Start")
	cli, err := clientv3.New(cliConfig)

	if err != nil {
		panic(err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = cli.Status(timeoutCtx, cliConfig.Endpoints[0])
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func ()  {
		
		defer wg.Done()
		time.Sleep(5*time.Second)
		cli.Put(context.Background(),"test_watch","ok")
	}()

	resChan := cli.Watch(context.Background(),"test_watch")
	wg.Add(1)
	go func () {
		defer wg.Done()
		fmt.Println("Start Watch")
		res := <- resChan
		fmt.Println(*res.Events[0])
	}()

	wg.Wait()

}