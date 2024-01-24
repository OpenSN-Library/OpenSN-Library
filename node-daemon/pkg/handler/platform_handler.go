package handler

import (
	"NodeDaemon/config"
	"NodeDaemon/model/ginmodel"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEtcdAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data: ginmodel.EtcdConfiguration{
			Address: config.GlobalConfig.Dependency.EtcdAddr,
			Port:    config.GlobalConfig.Dependency.EtcdPort,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

func GetRedisAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data: ginmodel.RedisConfiguration{
			Address:  config.GlobalConfig.Dependency.RedisAddr,
			Port:     config.GlobalConfig.Dependency.RedisPort,
			Index:    config.GlobalConfig.Dependency.RedisDBIndex,
			Password: config.GlobalConfig.Dependency.RedisPassword,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

func GetInfluxDBAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data: ginmodel.InfluxDBConfiguration{
			Address: config.GlobalConfig.Dependency.InfluxdbAddr,
			Port:    config.GlobalConfig.Dependency.InfluxdbPort,
			Token:   config.GlobalConfig.Dependency.InfluxdbToken,
			Enable:  config.GlobalConfig.App.EnableMonitor,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

func GetPlatformStatus(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "ok",
		Data:    nil,
	}
	ctx.JSON(http.StatusOK, resp)

}
