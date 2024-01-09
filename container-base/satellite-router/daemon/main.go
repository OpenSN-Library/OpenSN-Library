package main

import (
	"os"
	"os/signal"
	"satellite/data"
	"satellite/pkg/frr"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	err := frr.StartZebra()
	if err != nil {
		panic(err)
	}
	err = frr.StartOspfd()
	if err != nil {
		panic(err)
	}
	data.InitTopoInfoData()
	err = frr.InitConfigBatch()
	if err != nil {
		panic(err)
	}
	err = frr.WriteOspfConfig(frr.CommandBatchPath)
	if err != nil {
		panic(err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logrus.Infof("%v Recevied, Exit.", sig)
}
