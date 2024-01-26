package arranger

import (
	"NodeDaemon/model"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var nextIndex int = 1
var nextIndexLock *sync.Mutex = new(sync.Mutex)

func ArrangeLink(namespace *model.Namespace, instanceTarget map[int][]*model.InstanceConfig) (map[int][]*model.LinkConfig, error) {
	nextIndexLock.Lock()
	defer nextIndexLock.Unlock()
	getResp := utils.RedisClient.Get(
		context.Background(),
		key.NextLinkIndexKey,
	)
	if getResp.Err() != nil && getResp.Err() != redis.Nil {
		errMsg := fmt.Sprintf("Get Next Link Index Key Error: %s", getResp.Err().Error())
		logrus.Error(errMsg)
		return nil, getResp.Err()
	} else if getResp.Err() == redis.Nil {
		nextIndex = 1
	} else {
		nextIndex, _ = strconv.Atoi(getResp.Val())
	}
	actions := make(map[int][]*model.LinkConfig)
	nodeIndexSet := make(map[string]int)
	for index, v := range instanceTarget {
		for _, instance := range v {
			nodeIndexSet[instance.InstanceID] = index
		}
	}
	for link_index, linkInfo := range namespace.LinkConfig {

		namespace.LinkConfig[link_index].LinkIndex = nextIndex
		nextIndex = (nextIndex + 1) % (1 << 24)

		if linkInfo.EndInfos[0].InstanceID == "" {
			targetIndex1 := nodeIndexSet[linkInfo.EndInfos[1].InstanceID]
			namespace.LinkConfig[link_index].EndInfos[1].EndNodeIndex = targetIndex1
			actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])

		} else if linkInfo.EndInfos[1].InstanceID == "" {
			targetIndex0 := nodeIndexSet[linkInfo.EndInfos[0].InstanceID]
			namespace.LinkConfig[link_index].EndInfos[0].EndNodeIndex = targetIndex0
			actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])

		} else {
			targetIndex0 := nodeIndexSet[linkInfo.EndInfos[0].InstanceID]
			targetIndex1 := nodeIndexSet[linkInfo.EndInfos[1].InstanceID]
			namespace.LinkConfig[link_index].EndInfos[0].EndNodeIndex = targetIndex0
			namespace.LinkConfig[link_index].EndInfos[1].EndNodeIndex = targetIndex1

			logrus.Infof("Add Link Between %s and %s", linkInfo.EndInfos[0].InstanceID, linkInfo.EndInfos[1].InstanceID)
			if targetIndex0 == targetIndex1 {
				actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])
			} else {
				actions[targetIndex0] = append(actions[targetIndex0], &namespace.LinkConfig[link_index])
				actions[targetIndex1] = append(actions[targetIndex1], &namespace.LinkConfig[link_index])
			}
		}
	}
	setResp := utils.RedisClient.Set(
		context.Background(),
		key.NextLinkIndexKey,
		nextIndex,
		0,
	)
	if setResp.Err() != nil {
		errMsg := fmt.Sprintf("Set Next Link Index Key Error: %s", getResp.Err().Error())
		logrus.Error(errMsg)
	}
	return actions, nil
}
