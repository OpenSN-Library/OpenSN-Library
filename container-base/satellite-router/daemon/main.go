package main

import (
	"os"
	"os/signal"
	"satellite/pkg/frr"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	err := frr.StartFrr()
	if err != nil {
		panic(err)
	}
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logrus.Infof("%v Recevied, Exit.", sig)
}
