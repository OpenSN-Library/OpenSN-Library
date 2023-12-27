package data

import (
	"encoding/json"
	"os"
	"satellite/config"
	"satellite/model"

	"github.com/sirupsen/logrus"
)

var TopoInfo model.TopoInfo

func init() {
	success := false
	for !success {
		fd, err := os.Open(config.TopoInfoPath)
		if err != nil {
			logrus.Errorf("Open Topo Infomation File Error: %s", err.Error())
			panic(err)
		}
		decoder := json.NewDecoder(fd)
		err = decoder.Decode(&TopoInfo)
		if err != nil {
			logrus.Errorf("Decode Json Topo Infomation Error: %s", err.Error())
			panic(err)
		}
		success = true
	}
}
