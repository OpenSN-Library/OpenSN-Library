package main

import (
	"fmt"
	"os"
	"os/signal"
	"satellite/data"
	"satellite/pkg/configuration"
	"satellite/pkg/frr"
	"satellite/pkg/ifconfig"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	err := frr.StartOspfd()
	if err != nil {
		panic(err)
	}
	data.InitTopoInfoData()
	ifconfig.InitAddress()
	err = frr.InitConfigBatch()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 32; i++ {
		err = frr.WriteOspfConfig(frr.CommandBatchPath)
		if err != nil {
			fmt.Printf("Write Ospf Config Error: %s, Retry Time: %d, Max %d", err.Error(), i, 32)
		} else {
			break
		}
	}
	err = frr.StartZebra()
	if err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-sigChan:
			logrus.Infof("%v Recevied, Exit.", sig)
			return
		case newConfig := <-configuration.NewConfigurationChan:
			logrus.Infof("Configuration Modify Detected.")
			for k, v := range newConfig.LinkInfos {
				if _, ok := data.TopoInfo.LinkInfos[k]; !ok {
					data.TopoInfo.LinkInfos[k] = v
					ifconfig.SetInterfaceIP(k, v.V4Addr)
				} else if data.TopoInfo.LinkInfos[k].V4Addr != v.V4Addr {
					ifconfig.DelInterfaceIPs(k)
					ifconfig.SetInterfaceIP(k, v.V4Addr)
				}
			}
		}
	}

}
