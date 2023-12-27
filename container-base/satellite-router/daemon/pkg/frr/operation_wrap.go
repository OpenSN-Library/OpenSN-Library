package frr

import (
	"fmt"
	"os"
	"satellite/config"
	"satellite/data"

	"github.com/sirupsen/logrus"
)

var VtyConfig = `
frr version 7.2.1 
frr defaults traditional
hostname %s
log syslog informational
no ipv6 forwarding
service integrated-vtysh-config
!
router ospf
    redistribute connected

`

func init() {
	VtyConfig = fmt.Sprintf(VtyConfig, config.HostName)
	for i, v := range data.TopoInfo.LinkInfos {
		if data.TopoInfo.EndInfos[i].Type == config.Type {
			VtyConfig += fmt.Sprintf("\tnetwork %s area 0.0.0.0\n", v.V4Addr)
		}
	}
	VtyConfig += "!"
	confFile, err := os.Create("/etc/frr/frr.conf")
	if err != nil {
		logrus.Errorf("Create FRR Configuration Error: %s", err.Error())
	}
	_, err = confFile.Write([]byte(VtyConfig))
	if err != nil {
		logrus.Errorf("Write FRR Configuration Error: %s", err.Error())
	}
}
