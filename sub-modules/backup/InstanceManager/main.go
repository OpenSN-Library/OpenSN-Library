package main

import (
	"InstanceManager/biz/service"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func signalHandler(sigChan chan os.Signal, modules []service.Module) {
	sig := <-sigChan
	infoMsg := fmt.Sprintf("Signal %v Detected Gracefully Dying...", sig)
	logrus.Info(infoMsg)
	for _, v := range modules {
		v.Stop()
	}
}

func main() {
	modules := []service.Module{
		service.CreateInstanceWatchModule(),
	}

	for _, v := range modules {
		v.Run()
	}

	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGTERM, syscall.SIGINT)
	go signalHandler(sysSigChan, modules)
	for _, v := range modules {
		v.Wait()
	}
}
