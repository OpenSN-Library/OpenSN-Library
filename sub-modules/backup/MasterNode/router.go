package main

import (
	"MasterNode/biz/handler"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(r *gin.Engine) {
	api := r.Group("/api")
	platform := api.Group("/platform")
	platform.GET("/address/etcd", handler.GetEtcdAddressHandler)
	platform.GET("/address/redis", handler.GetRedisAddressHandler)
	platform.GET("/status", handler.GetPlatformStatus)
	namespace := api.Group("/namespace")
	namespace.GET("/list", handler.GetNsListHandler)
	namespace.POST("/create", handler.CreateNsHandler)
	namespace.POST("/:name/update", handler.UpdateNsHandler)
	namespace.POST("/:name/start", handler.StartNsHandler)
	namespace.POST("/:name/stop", handler.StopNsHandler)
	namespace.DELETE("/:name", handler.DeleteNsHandler)
	namespace.GET("/:name", handler.GetNodeInfoHandler)
}
