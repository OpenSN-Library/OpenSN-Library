package main

import (
	"NodeDaemon/biz"
	"NodeDaemon/config"
	"NodeDaemon/utils/tools"
	"time"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func signalHandler(sigChan chan os.Signal, modules []biz.Module) {
	sig := <-sigChan
	infoMsg := fmt.Sprintf("Signal %v Detected Gracefully Dying...", sig)
	logrus.Info(infoMsg)
	for _, v := range modules {
		v.Stop()
	}
}

func main() {

	modules := []biz.Module{
		biz.CreateInstanceModuleTask(),
		biz.CreateLinkModuleTask(),
		biz.CreateStatusUpdateModule(),
	}
	if config.StartMode == config.MasterNode {
		masterNodeModule := biz.CreateMasterNodeModuleTask()
		masterNodeModule.Run()
		tools.WaitSuccess(masterNodeModule.IsSetupFinish, 3*time.Minute, 10*time.Second)
		logrus.Info("Master Node Init Success.")
		modules = append(modules, masterNodeModule)
	}
	err := biz.NodeInit()
	if err != nil {
		errMsg := fmt.Sprintf("Init Node Error: %s", err.Error())
		logrus.Error(errMsg)
		panic(err)
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
