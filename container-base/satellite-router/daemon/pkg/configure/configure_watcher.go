package configure

import (
	"encoding/json"
	"fmt"
	"os"
	"satellite/data"
	"satellite/model"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

var NewConfigurationChan chan model.TopoInfo = make(chan model.TopoInfo)
var ConfigureFileWatcher *fsnotify.Watcher

func parseConfigFile(path string) (model.TopoInfo, error) {
	var ret model.TopoInfo
	fd, err := os.Open(path)
	defer fd.Close()
	if err != nil {
		logrus.Errorf("Open Topo Infomation File Error: %s", err.Error())
		return ret, err
	}
	decoder := json.NewDecoder(fd)
	err = decoder.Decode(&ret)
	if err != nil {
		logrus.Errorf("Decode Json Topo Infomation Error: %s", err.Error())
		return ret, err
	}
	return ret, nil
}

func InitConfigurationWatcher(watchPath string) {
	wather, err := fsnotify.NewWatcher()
	if err != nil {
		err = fmt.Errorf("create configuration file watcher error: %s", err.Error())
		panic(err)
	}
	wather.Add(watchPath)
	go func() {
		defer wather.Close()
		for event := range wather.Events {
			if event.Op&fsnotify.Write == fsnotify.Write {
				newConfiguration, err := parseConfigFile(watchPath)
				if err == nil {
					data.TopoInfo = newConfiguration
				}
			}
		}
	}()
}
