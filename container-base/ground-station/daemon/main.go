package main

import (
	"ground/config"
	"ground/data"
	"ground/pkg/configure"
	"ground/pkg/ifconfig"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	data.InitTopoInfoData()
	ifconfig.InitInterfaceWatcher()
	configure.InitConfigurationWatcher(config.TopoInfoPath)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
