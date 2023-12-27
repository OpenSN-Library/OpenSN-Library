package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const Type = "Satellite"

var TopoInfoPath = "/share/topo/%s.json"

var HostName string

func init() {
	hostnameEnv, err := os.Hostname()
	if err != nil {
		logrus.Errorf("Get Hostname Error: %s", err.Error())
		panic(err)
	}
	HostName = hostnameEnv
	TopoInfoPath = fmt.Sprintf(TopoInfoPath, HostName)
}
