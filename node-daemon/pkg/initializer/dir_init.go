package initializer

import (
	"NodeDaemon/share/dir"
	"NodeDaemon/utils"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

func InitWorkdir() error {
	wd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Get Workdir Dir Error: %s", err.Error())
		return err
	}
	dir.WorkDirRoot = path.Join(wd, "runtime")
	dir.MountShareData = path.Join(dir.WorkDirRoot, "share")
	dir.TopoInfoDir = path.Join(dir.MountShareData, "topo")
	dir.UserDataDir = path.Join(dir.MountShareData, "user")
	err = utils.CreateDirNX(dir.WorkDirRoot)
	if err != nil {
		logrus.Errorf("Create Runtime Dir %s Error: %s", dir.WorkDirRoot, err.Error())
		return err
	}
	err = utils.CreateDirNX(dir.MountShareData)
	if err != nil {
		logrus.Errorf("Create Runtime Dir %s Error: %s", dir.MountShareData, err.Error())
		return err
	}
	err = utils.CreateDirNX(dir.TopoInfoDir)
	if err != nil {
		logrus.Errorf("Create Runtime Dir %s Error: %s", dir.TopoInfoDir, err.Error())
		return err
	}
	err = utils.CreateDirNX(dir.UserDataDir)
	if err != nil {
		logrus.Errorf("Create Runtime Dir %s Error: %s", dir.UserDataDir, err.Error())
		return err
	}
	return nil
}
