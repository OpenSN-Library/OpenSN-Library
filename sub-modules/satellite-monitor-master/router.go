package main

import (
	"github.com/gin-gonic/gin"
	"satellite/monitor/handler"
	"satellite/monitor/middleware"
)

func register(r *gin.Engine) {
	r.Use(middleware.CORS)
	r.GET("/api/satellite/list", handler.GetConnectionInfo)
	r.POST("/api/satellite/list", handler.SetConnectionInfo) // called by set_monitor
	r.GET("/api/command/traceroute", handler.GetTracerouteResultHandler)
	r.GET("/api/satellite/status", handler.GetLinkState)
	r.GET("/api/satellite/interfaces", handler.GetInterfacesHandler)
	r.POST("/api/satellite/print", handler.PrintHandler)
	r.POST("/api/ground/position", handler.SetGroundStationPositionHandler)
	r.GET("/api/ground/position", handler.GetGroundStationPositionHandler)
	r.GET("/api/ground/connection", handler.GetGroundStationConnHandler)
	r.GET("/api/video/transition", handler.TransitionVideo)
}
