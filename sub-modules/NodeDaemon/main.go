package main

import (
	"NodeDaemon/biz"
	"NodeDaemon/utils"
	"strings"
	"time"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func signalHandler(sigChan chan os.Signal, modules []biz.Module) {
	<-sigChan
	for _, v := range modules {
		v.Stop()
	}
}

func parseArgs(args []string) (ret biz.Parameter) {
	for _, v := range args {
		v = strings.TrimLeft(v, "-")
		kv := strings.Split(v, "=")
		if len(kv) < 2 {
			continue
		}
		switch kv[0] {
		case "master":
			ret.MasterNodeAddr = kv[1]
		case "interface":
			ret.BindInterfaceName = kv[1]
		case "mode":
			if kv[1] != biz.MasterNode && kv[1] != biz.ServantNode {
				errMsg := fmt.Sprintf("Invalid Mode %s, expected master or servant", kv[1])
				logrus.Error(errMsg)
				panic(errMsg)
			} else {
				ret.NodeMode = kv[1]
			}
		}
	}
	return
}

func main() {
	var para biz.Parameter
	if len(os.Args) <= 1 || os.Args[1] == "help" {
		goto help
	} else if os.Args[1] == "start" {
		if len(os.Args) <= 2 {
			para.NodeMode = biz.MasterNode
			goto normal_start
		}
		para = parseArgs(os.Args[2:])
		goto normal_start
	}

help:

normal_start:
	modules := []biz.Module{
		biz.CreateInstanceModuleTask(),
		biz.CreateLinkModuleTask(),
		biz.CreateStatusUpdateModule(),
	}
	if para.NodeMode == biz.MasterNode {
		masterNodeModule := biz.CreateMasterNodeModuleTask()
		masterNodeModule.Run()
		utils.WaitSuccess(masterNodeModule.IsSetupFinish, 3*time.Minute, 10*time.Second)
		modules = append(modules, masterNodeModule)
	}
	err := biz.NodeInit(para)
	if err != nil {
		errMsg := fmt.Sprintf("Init Node Error: %s", err.Error())
		logrus.Error(errMsg)
		panic(err)
	}

	for _, v := range modules {
		v.Run()
	}

	sysSigChan := make(chan os.Signal, 1)
	signal.Notify(sysSigChan, syscall.SIGTERM)
	go signalHandler(sysSigChan, modules)
	for _, v := range modules {
		v.Wait()
	}
}
