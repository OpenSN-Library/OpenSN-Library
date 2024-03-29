package handler

import (
	"NodeDaemon/model/ginmodel"
	"NodeDaemon/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetItemsWithPrefixHandler(ctx *gin.Context) {
	respData := map[string]interface{}{}
	prefix := ctx.Query("prefix")
	if prefix == "" {
		prefix = "/"
	}
	resp, err := utils.EtcdClient.Get(
		context.Background(),
		prefix,
		clientv3.WithPrefix(),
	)

	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Get Database Items Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}

	for _, kv := range resp.Kvs {
		respData[string(kv.Key)] = string(kv.Value)
	}
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
		Data:    respData,
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func UpdateDatabaseItemHandler(ctx *gin.Context) {
	var req ginmodel.UpdateDatabaseItemRequest
	if err := ctx.Bind(&req); err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Request Data: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	_, err := utils.EtcdClient.Put(
		context.Background(),
		req.Key,
		req.Val,
	)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Update Database Item Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, jsonResp)
}

func DeleteDatabaseItemHandler(ctx *gin.Context) {
	var req ginmodel.DeleteDatabaseItemRequest
	if err := ctx.Bind(&req); err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Invalid Request Data: %s", err.Error()),
		}
		ctx.JSON(http.StatusBadRequest, jsonResp)
		return
	}
	_, err := utils.EtcdClient.Delete(
		context.Background(),
		req.Key,
		clientv3.WithPrefix(),
	)
	if err != nil {
		jsonResp := ginmodel.JsonResp{
			Code:    -1,
			Message: fmt.Sprintf("Delete Database Item Error: %s", err.Error()),
		}
		ctx.JSON(http.StatusInternalServerError, jsonResp)
		return
	}
	jsonResp := ginmodel.JsonResp{
		Code:    0,
		Message: "Success",
	}
	ctx.JSON(http.StatusOK, jsonResp)
}
