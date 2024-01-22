package handler

import (
	"NodeDaemon/model"
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/share/data"
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

func GetNamespaceInstancePostion(ctx *gin.Context) {
	spaceName := ctx.Param("name")
	var retData = map[string]model.Position{}
	info, ok := data.NamespaceMap[spaceName]
	if !ok {
		errMsg := fmt.Sprintf("Namespace %s Not Found.", spaceName)
		logrus.Error(errMsg)
		resp := ginmodel.JsonResp{
			Code:    -1,
			Message: errMsg,
		}
		ctx.JSON(http.StatusNotFound, resp)
		return
	}
	if !info.Running {
		errMsg := fmt.Sprintf("Namespace %s is not Running.", spaceName)
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
		fmt.Sprintf(key.NamespaceInstancePositionTemplate, spaceName),
		clientv3.WithPrefix(),
	)
	if err != nil {
		errMsg := fmt.Sprintf("Get Postion Data of Namespace %s Error: %s", spaceName, err.Error())
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
	for _, v := range info.InstanceConfig {
		if _, ok := retData[v.InstanceID]; !ok {
			retData[v.InstanceID] = model.Position{}
		}
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data:    retData,
	}
	ctx.JSON(http.StatusOK, resp)
}
