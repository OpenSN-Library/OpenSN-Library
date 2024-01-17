package main

import (
	"fmt"
	"os"
	"os/signal"
	"satellite/data"
	"satellite/pkg/frr"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	err := frr.StartOspfd()
	if err != nil {
		panic(err)
	}
	data.InitTopoInfoData()
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
	sig := <-sigChan
	logrus.Infof("%v Recevied, Exit.", sig)
}
