package main

import (
	"NodeDaemon/config"
	"NodeDaemon/pkg/initializer"
	"NodeDaemon/pkg/module"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"

	"github.com/sirupsen/logrus"
)

const DefaultConfigPath = "config/config.json"

func signalHandler(sigChan chan os.Signal, modules []module.Module) {
	sig := <-sigChan
	infoMsg := fmt.Sprintf("Signal %v Detected Gracefully Dying...", sig)
	logrus.Info(infoMsg)
	for _, v := range modules {
		v.Stop()
	}
}

func main() {

	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
	})

	if len(os.Args) > 2 {
		config.InitConfig(os.Args[1])
	} else {
		config.InitConfig(DefaultConfigPath)
	}
	
	err := initializer.NodeInit()
	if err != nil {
		errMsg := fmt.Sprintf("Init Node Error: %s", err.Error())
		logrus.Error(errMsg)
		panic(err)
	}
	modules := []module.Module{
		module.CreateInstanceModuleTask(),
		module.CreateLinkModuleTask(),
		module.CreateStatusUpdateModule(),
		module.CreateHealthyCheckTask(),
		module.CreateNodeWatchTask(),
	}
	if !config.GlobalConfig.App.IsServant {
		masterNodeModule := module.CreateMasterNodeModuleTask()
		modules = append(modules, masterNodeModule)
	}

	for _, v := range modules {
		v.Run()
	}

	go func() {
		for _, ch := range modules {
			err := ch.CheckError()
			if err != nil {
				errMsg := fmt.Sprintf("Sub Rountine Error:%s", err.Error())
				logrus.Error(errMsg)
			}
		}
		time.Sleep(2 * time.Second)
	}()

	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGTERM, syscall.SIGINT)
	go signalHandler(sysSigChan, modules)
	for _, v := range modules {
		v.Wait()
	}
}
