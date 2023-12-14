package handler

import (
	"MasterNode/config"
	"MasterNode/model/ginmodel"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEtcdAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code: 0,
		Message: "success",
		Data: ginmodel.EtcdConfiguration {
			Address: config.EtcdAddr,
			Port: config.EtcdPort,
		},
	}
	ctx.JSON(http.StatusOK,resp)
}

func GetRedisAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code: 0,
		Message: "success",
		Data: ginmodel.RedisConfiguration {
			Address: config.RedisAddr,
			Port: config.RedisPort,
			Index: config.RedisDBIndex,
			Password: config.RedisPassword,
		},
	}
	ctx.JSON(http.StatusOK,resp)
}

func GetPlatformStatus(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code: 0,
		Message: "ok",
		Data: nil,
	}
	ctx.JSON(http.StatusOK,resp)

}