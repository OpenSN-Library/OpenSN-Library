package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"github.com/sirupsen/logrus"
)

type WebShellInfo struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
	Pid  int    `json:"pid"`
}

func AllocTcpPort(addr []byte) (int, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", FormatIPv4(addr)))
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func StartWebShell(addr string, port int, writeable bool, cmd string, args []string, timeoutMin int, callback func()) (WebShellInfo, error) {
	finalCmd := []string{
		"./gotty",
		"-a", addr,
		"-p", strconv.Itoa(port),
		"--timeout", strconv.Itoa(60 * timeoutMin),
		"--title-format", "WebShell",
	}
	if writeable {
		finalCmd = append(finalCmd, "-w")
	}
	finalCmd = append(finalCmd, cmd)
	finalCmd = append(finalCmd, args...)
	execCmd := exec.Command(finalCmd[0], finalCmd[1:]...)
	err := execCmd.Start()
	if err != nil {
		return WebShellInfo{}, fmt.Errorf("start webShell error: %s", err.Error())
	}
	go func() {
		execCmd.Wait()
		if err != nil {
			logrus.Errorf("WebShell Running Error: %s", err.Error())
		}
	}()
	return WebShellInfo{
		Addr: addr,
		Port: port,
		Pid:  execCmd.Process.Pid,
	}, nil
}

func KillWebShell(pid int) error {
	execCmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	err := execCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
