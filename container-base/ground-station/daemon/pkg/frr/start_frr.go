package frr

import (
	"os/exec"

	"github.com/sirupsen/logrus"
)

func StartFrr() error {
	// service frr start
	cmd := exec.Command(
		"service",
		"frr",
		"start",
	)
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Start Zebra Error: %s", err.Error())
		return err
	}
	return err
}

func StartZebra() error {
	// /usr/lib/frr/zebra -d -F traditional -A 127.0.0.1 -s 90000000
	cmd := exec.Command(
		"/usr/lib/frr/zebra", "-d",
		"-F", "traditional",
		"-A", "127.0.0.1",
		"-s", "90000000",
	)
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Start Zebra Error: %s", err.Error())
		return err
	}
	return err
}

func StartOspfd() error {
	// /usr/lib/frr/ospfd -d -F traditional -A 127.0.0.1
	cmd := exec.Command(
		"/usr/lib/frr/ospfd", "-d",
		"-F", "traditional",
		"-A", "127.0.0.1",
	)
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Start Ospfd Error: %s", err.Error())
		return err
	}
	return err
}

func WriteOspfConfig(batchPath string) error {
	cmd := exec.Command(
		"/usr/bin/vtysh",
		"-f", batchPath,
	)
	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Config Ospfd Error: %s", err.Error())
		return err
	}
	return err
}
