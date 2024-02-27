package data

import (
	"encoding/json"
	"ground/config"
	"ground/model"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var TopoInfo model.TopoInfo

func InitTopoInfoData() {
retry:
	fd, err := os.Open(config.TopoInfoPath)
	defer fd.Close()
	if err != nil {
		logrus.Errorf("Open Topo Infomation File Error: %s", err.Error())
		time.Sleep(time.Second)
		goto retry
	}
	decoder := json.NewDecoder(fd)
	err = decoder.Decode(&TopoInfo)
	if err != nil {
		logrus.Errorf("Decode Json Topo Infomation Error: %s", err.Error())
		return
	}

}
