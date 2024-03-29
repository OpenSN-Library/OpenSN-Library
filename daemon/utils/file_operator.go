package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func CreateDirNX(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0666)
		if err != nil {
			return fmt.Errorf("create dir error %s", err.Error())
		}
		return nil
	}
	if file.IsDir() {
		return nil
	} else {
		return fmt.Errorf("%s exists but is not dir", path)
	}

}

func CreateFileNX(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		fd, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create file error %s", err.Error())
		}
		fd.Close()
		return nil
	}
	if file.IsDir() {
		return nil
	} else {
		return fmt.Errorf("%s exists but is dir", path)
	}
}

func WriteToFile(path string, data []byte) error {
	fd, err := os.Create(path)

	if err != nil {
		return err
	}

	fd.Write(data)

	return fd.Close()
}

func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		logrus.Errorf("Delete File %s Error: %s", path, err.Error())
	}
	return err
}
