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
	platform.GET("/address/redis", handler.GetRedisAddressHandler)
	platform.GET("/address/influxdb", handler.GetInfluxDBAddressHandler)
	platform.GET("/status", handler.GetPlatformStatus)
	namespace := api.Group("/namespace")
	namespace.GET("/list", handler.GetNsListHandler)
	namespace.POST("/create", handler.CreateNsHandler)
	namespace.POST("/:name/start", handler.StartNsHandler)
	namespace.POST("/:name/stop", handler.StopNsHandler)
	namespace.DELETE("/:name", handler.DeleteNsHandler)
	namespace.GET("/:name", handler.GetNsInfoHandler)
	position := api.Group("/position")
	position.GET("/:name", handler.GetNamespaceInstancePostion)
	node := api.Group("/node")
	node.GET("/list", handler.GetNodeListHandler)
	node.GET("/:node_index", handler.GetNodeInfoHandler)

}

func RegisterStatics(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) { // 当 API 不存在时，返回静态文件
		path := c.Request.URL.Path                                          // 获取请求路径
		s := strings.Split(path, ".")                                       // 分割路径，获取文件后缀
		prefix := "ui"                                                      // 前缀路径
		if data, err := static.Static.ReadFile(prefix + path); err != nil { // 读取文件内容
			// 如果文件不存在，返回首页 index.html
			if data, err = static.Static.ReadFile(prefix + "/index.html"); err != nil {
				c.JSON(404, gin.H{
					"err": err,
				})
			} else {
				c.Data(200, mime.TypeByExtension(".html"), data)
			}
		} else {
			// 如果文件存在，根据请求的文件后缀，设置正确的mime type，并返回文件内容
			c.Data(200, mime.TypeByExtension(fmt.Sprintf(".%s", s[len(s)-1])), data)
		}
	})
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
