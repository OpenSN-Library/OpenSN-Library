package main

import (
	"MasterNode/biz/handler"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(r *gin.Engine) {
	platform := r.Group("/platform")
	platform.GET("/address/etcd", handler.GetEtcdAddressHandler)
	platform.GET("/address/redis", handler.GetRedisAddressHandler)
	platform.GET("/status", handler.GetPlatformStatus)
}
