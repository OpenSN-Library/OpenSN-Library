package module

import (
	"NodeDaemon/config"
	"NodeDaemon/pkg/handler"
	"NodeDaemon/share/signal"
	"NodeDaemon/share/static"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const MasterNodeContainerName = "master_node"

type MasterNodeModule struct {
	Base
}

func RegisterHandlers(r *gin.Engine) {
	api := r.Group("/api")
	platform := api.Group("/platform")
	platform.GET("/address/etcd", handler.GetEtcdAddressHandler)
	platform.GET("/address/influxdb", handler.GetInfluxDBAddressHandler)
	platform.GET("/status", handler.GetPlatformStatus)

	emulate := api.Group("/emulation")
	emulate.POST("/update", handler.ConfigEmulationHandler)
	emulate.POST("/start", handler.StartEmulationHandler)
	emulate.POST("/stop", handler.StopEmulationHandler)
	emulate.GET("/", handler.GetEmulationConfigHandler)
	emulate.POST("/topology", handler.AddTopologyHandler)
	emulate.POST("/reset", handler.ResetStatusHandler)

	position := api.Group("/position")
	position.GET("/:instance_id", handler.GetInstancePositionHandler)

	node := api.Group("/node")
	node.GET("/", handler.GetNodeListHandler)
	node.GET("/:node_index", handler.GetNodeInfoHandler)

	instance := api.Group("/instance")
	instance.GET("/", handler.GetInstanceListHandler)
	instance.GET("/:node_index/:instance_id", handler.GetInstanceInfoHandler)
	instance.POST("/start", handler.StartInstanceHander)
	instance.POST("/stop", handler.StopInstanceHandler)
	instance.POST("/:instance_id", handler.AddInstanceHandler)
	instance.DELETE("/:instance_id", handler.DelInstanceHandler)

	link := api.Group("/link")
	link.GET("/", handler.GetLinkListHandler)
	link.GET("/:node_index/:link_id", handler.GetLinkInfoHandler)
	link.POST("/", handler.AddLinkHandler)
	link.DELETE("/:link_id", handler.DelLinkHandler)

	linkParameter := api.Group("/link_parameter")
	linkParameter.GET("/", handler.GetLinkParameterListHandler)
	linkParameter.GET("/:link_id", handler.GetLinkParameterHandler)
	linkParameter.POST("/:link_id", handler.UpdateLinkParameterHandler)

	database := api.Group("/database")
	database.GET("/items", handler.GetItemsWithPrefixHandler)
	database.POST("/update", handler.UpdateDatabaseItemHandler)
	database.POST("/delete", handler.DeleteDatabaseItemHandler)

	resourceLast := api.Group("/resource/last")
	resourceLast.GET("/node/:node_index", handler.GetNodeResourceDataHandler)
	resourceLast.GET("/instance/:instance_id", handler.GetInstanceResourceDataHandler)
	resourceLast.GET("/link/:link_id", handler.GetLinkResourceDataHander)

	resourcePeriod := api.Group("/resource/period")
	resourcePeriod.GET("/node/:node_index", handler.GetPeriodNodeResourceDataHandler)
	resourcePeriod.GET("/instance/:instance_id", handler.GetPeriodInstanceResourceDataHandler)
	resourcePeriod.GET("/link/:link_id", handler.GetPeriodLinkResourceDataHander)

	webshell := api.Group("/webshell")
	webshell.POST("/instance", handler.StartInstanceWebshellHandler)
	webshell.POST("/link", handler.StartLinkWebshellHandler)
}

func RegisterStatics(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) { // 当 API 不存在时，返回静态文件
		path := c.Request.URL.Path
		s := strings.Split(path, ".")
		prefix := "ui"
		if data, err := static.Static.ReadFile(prefix + path); err != nil {
			if data, err = static.Static.ReadFile(prefix + "/index.html"); err != nil {
				c.JSON(404, gin.H{
					"err": err,
				})
			} else {
				c.Data(200, mime.TypeByExtension(".html"), data)
			}
		} else {
			c.Data(200, mime.TypeByExtension(fmt.Sprintf(".%s", s[len(s)-1])), data)
		}
	})
}

func masterDaemonFunc(sigChann chan int, errChann chan error) {
	r := gin.Default()
	if !config.GlobalConfig.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r.Use(cors.Default())
	RegisterHandlers(r)
	RegisterStatics(r)
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", config.GlobalConfig.App.ListenPort),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChann <- err
			logrus.Error("HTTP Server Dead: ", err.Error())
		}
	}()

	for {
		sig := <-sigChann
		if sig == signal.STOP_SIGNAL {
			return
		}
	}
}

func CreateMasterNodeModuleTask() *MasterNodeModule {
	return &MasterNodeModule{
		Base{
			sigChan:    make(chan int),
			errChan:    make(chan error),
			running:    false,
			daemonFunc: masterDaemonFunc,
			wg:         new(sync.WaitGroup),
			ModuleName: "MasterNode",
		},
	}
}
