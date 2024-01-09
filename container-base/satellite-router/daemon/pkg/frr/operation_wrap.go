package frr

import (
	"fmt"
	"os"
	"satellite/config"
	"satellite/data"

	"github.com/sirupsen/logrus"
)

var CommandBatch = `
router ospf
    redistribute connected

`
var CommandBatchPath = "/etc/frr/batch.txt"

func InitConfigBatch() error {
	for i, v := range data.TopoInfo.LinkInfos {
		if data.TopoInfo.EndInfos[i].Type == config.Type {
			CommandBatch += fmt.Sprintf("\tnetwork %s area 0.0.0.0\n", v.V4Addr)
		}
	}
	confFile, err := os.Create(CommandBatchPath)
	if err != nil {
		logrus.Errorf("Create FRR Configuration Error: %s", err.Error())
		return err
	}
	_, err = confFile.Write([]byte(CommandBatch))
	if err != nil {
		logrus.Errorf("Write FRR Configuration Error: %s", err.Error())
		return err
	}
	return nil
}
