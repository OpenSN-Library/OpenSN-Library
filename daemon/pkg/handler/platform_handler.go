package handler

import (
	"NodeDaemon/config"
	"NodeDaemon/model/ginmodel"
	"net/http"
	"time"

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

func GetCodeServerAddress(ctx *gin.Context) {
	if config.GlobalConfig.Dependency.CodeServerAddr == "" {
		resp := ginmodel.JsonResp{
			Code:    0,
			Message: "success",
			Data: ginmodel.CodeServerConfiguration{
				Address: "",
				Port:    0,
			},
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}
	resp := ginmodel.JsonResp{
		Code:    0,
		Message: "success",
		Data: ginmodel.CodeServerConfiguration{
			Address: config.GlobalConfig.Dependency.CodeServerAddr,
			Port:    config.GlobalConfig.Dependency.CodeServerPort,
		},
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetUnixTimestampMillis is a handler function that returns the current Unix timestamp in milliseconds.
// For synchronous calls, the function returns the current Unix timestamp in milliseconds.
// When call this interface, please record the cost of the request and response time.
// And set the system time to the response time plus half of the request-response time.
func GetUnixTimestampMillis(ctx *gin.Context) {
	ctx.String(http.StatusOK, "%d", time.Now().UnixMicro())
}
