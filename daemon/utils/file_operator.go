package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	DirType    = 1
	BinaryType = 2
	TextType   = 3
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

func CheckPathType(path string) (int, error) {
	file, err := os.Stat(path)
	if err != nil {

	}
	if file.IsDir() {
		return DirType, nil
	}
	fd, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer fd.Close()
	buf := make([]byte, 8192)
	n, err := fd.Read(buf)
	if err != nil {
		return 0, err
	}
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return BinaryType, nil
		}
	}
	return TextType, nil
}
