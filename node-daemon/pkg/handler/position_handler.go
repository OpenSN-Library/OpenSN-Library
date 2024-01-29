package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/pkg/synchronizer"
	"NodeDaemon/share/key"
	"NodeDaemon/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetInstancePostion(ctx *gin.Context) {
	emulation_config, err := synchronizer.GetEmulationInfo()

	if err != nil {
		errMsg := fmt.Sprintf("Get Emulation Info Error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	var retData = map[string]model.Position{}
	if !emulation_config.Running {
		errMsg := "Emulation is not Running."
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	positions, err := utils.EtcdClient.Get(
		context.Background(),
		fmt.Sprintf(key.NamespaceInstancePositionTemplate),
		clientv3.WithPrefix(),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Get Postion Data Error: %s", err.Error())
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	for _, v := range positions.Kvs {
		var positionMap = map[string]model.Position{}
		err := json.Unmarshal(v.Value, &positionMap)
		if err != nil {
			errMsg := fmt.Sprintf("Parse Position Data of %s Error: %s", v.Key, err.Error())
			logrus.Error(errMsg)
			continue
		}
		for k, v := range positionMap {
			retData[k] = v
		}
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    retData,
	}
	ctx.JSON(http.StatusOK, resp)
}
