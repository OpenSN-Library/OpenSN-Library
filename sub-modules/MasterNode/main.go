package main

import (
	"MasterNode/biz/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {

	modules := []service.Module{
		service.CreateEtcdModuleTask(),
		service.CreateRedisModuleTask(),
		service.CreateNodeWatchTask(),
		service.CreateHealthyCheckTask(),
	}

	for _, v := range modules {
		v.Run()
	}

	r := gin.Default()
	RegisterHandlers(r)
	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Error("HTTP Server Dead: ", err.Error())
		}
	}()
	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGTERM)

	go func() {
		<-sysSigChan
		srv.Close()
		for _, v := range modules {
			v.Stop()
		}
	}()

	for _, v := range modules {
		v.Wait()
	}
}
