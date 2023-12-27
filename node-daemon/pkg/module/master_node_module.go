package module

import (
	"NodeDaemon/config"
	"NodeDaemon/pkg/handler"
	"NodeDaemon/share/signal"
	"errors"
	"net/http"
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

func masterDaemonFunc(sigChann chan int, errChann chan error) {
	// logger := logrus.New()
	// logger.SetFormatter(&nested.Formatter{
	// 	TimestampFormat: time.RFC3339,
	// })
	r := gin.Default()
	if !config.GlobalConfig.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r.Use(cors.Default())
	RegisterHandlers(r)
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
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
