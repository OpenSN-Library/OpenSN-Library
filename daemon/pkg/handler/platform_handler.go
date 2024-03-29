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

func GetInfluxDBAddressHandler(ctx *gin.Context) {
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data: ginmodel.InfluxDBConfiguration{
			Address: config.GlobalConfig.Dependency.InfluxdbAddr,
			Bucket:  config.GlobalConfig.Dependency.InfluxdbBucket,
			Port:    config.GlobalConfig.Dependency.InfluxdbPort,
			Org:     config.GlobalConfig.Dependency.InfluxdbOrg,
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
