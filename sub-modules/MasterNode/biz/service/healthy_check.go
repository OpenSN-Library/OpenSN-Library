package service

import (
	"MasterNode/data"
	"MasterNode/utils"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var maxBeatSecond = 40
var checkGap = maxBeatSecond / 2 * time.Now().Second()

type HealthyCheckModule struct {
	ModuleBase
}

func CreateHealthyCheckTask() *HealthyCheckModule {
	return &HealthyCheckModule{
		ModuleBase{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			wg:         new(sync.WaitGroup),
			daemonFunc: checkNodeHealthy,
			runing:     false,
		},
	}
}

func updateNodeList(addSet map[int]bool, delSet map[int]bool) error {

	if len(addSet) <= 0 && len(delSet) <= 0 {
		return nil
	}

	var remoteIndexList []int = []int{}
	var indexes2Update []int = []int{}
	status := utils.LockKeyWithTimeout(data.NodeIndexListKey, 6*time.Second)
	if !status {
		return fmt.Errorf("unable to access lock of key %s", data.NodeIndexListKey)
	}
	getResp, err := utils.EtcdClient.Get(context.Background(), data.NodeIndexListKey)
	if err != nil {
		return err
	}

	if len(getResp.Kvs) >= 1 {
		err = json.Unmarshal(getResp.Kvs[0].Value, &remoteIndexList)

		if err != nil {
			return err
		}
	}

	for _, v := range remoteIndexList {
		if !delSet[v] && !addSet[v] {
			indexes2Update = append(indexes2Update, v)
		}
	}

	for k, v := range addSet {
		if v {
			indexes2Update = append(indexes2Update, k)
		}
	}

	updateBytes, err := json.Marshal(indexes2Update)

	if err != nil {
		return err
	}

	_, err = utils.EtcdClient.Put(context.Background(), data.NodeIndexListKey, string(updateBytes))

	return err
}

func checkNodeHealthy(sigChan chan int, errChan chan error) {

	for {
		if utils.CheckRedisServe() {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	for {

		select {
		case sig := <-sigChan:
			if sig == data.STOP_SIGNAL {
				return
			}
		default:
			// DO NOTHING, JUST FOR NO BLOCKING
		}

		delSet := make(map[int]bool)
		redisResp := utils.RedisClient.HGetAll(context.Background(), data.NodeHeartBeatKey)
		if redisResp.Err() != nil {
			err := redisResp.Err()
			logrus.Error("Check Node Healthy Error: ", err.Error())
		}
		currentTime := time.Now()
		for k, v := range redisResp.Val() {
			timeUnix, err := strconv.ParseInt(v, 10, 64)
			nodeIndex, _ := strconv.Atoi(k)
			if err != nil {
				lastBeat := time.Unix(timeUnix, 0)
				delta := currentTime.Sub(lastBeat)
				if delta.Seconds() > float64(maxBeatSecond) {
					if nodeIndex != 0 {
						delSet[nodeIndex] = true
					}
				}
			}
		}

		err := utils.DoWithRetry(func() error {
			return updateNodeList(map[int]bool{}, delSet)
		}, 1)

		if err != nil {
			logrus.Error("Delete Dead Node Failed: ", err.Error())
		}

		time.Sleep(time.Duration(checkGap))
	}
}
