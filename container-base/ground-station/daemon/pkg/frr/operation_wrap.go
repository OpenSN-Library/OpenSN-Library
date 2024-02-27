package frr

import (
	"fmt"
	"ground/model"
	"os"

	"github.com/sirupsen/logrus"
)

var CommandBatch = `
router ospf
    redistribute connected
`
var CommandBatchPath = "/etc/frr/batch.txt"

func InitConfigBatch(links []*model.LinkInfo) error {
	batch := CommandBatch
	for _, v := range links {
		batch += fmt.Sprintf("\tnetwork %s area 0.0.0.0\n", v.V4Addr)
	}
	confFile, err := os.Create(CommandBatchPath)
	defer confFile.Close()
	if err != nil {
		logrus.Errorf("Create FRR Configuration Error: %s", err.Error())
		return err
	}
	_, err = confFile.Write([]byte(batch))
	if err != nil {
		logrus.Errorf("Write FRR Configuration Error: %s", err.Error())
		return err
	}
	err = WriteOspfConfig(CommandBatchPath)

	if err != nil {
		logrus.Errorf("Write FRR Configuration to VTY Error: %s", err.Error())
		return err
	}
	return nil
}
