package frr

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

func StartFrr() error {
	cmd := exec.Command("service", "frr", "start")
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Start Frr Error: %s", err.Error())
		return err
	}
	return err
}
