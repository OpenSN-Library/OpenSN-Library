package ifconfig

import (
	"ground/data"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const checkGap = 500 * time.Millisecond

var linkReadymap = make(map[string]bool)

func InitInterfaceWatcher() {
	go InterfaceWatcher()
}

func InterfaceWatcher() {
	for {
		links, err := netlink.LinkList()
		if err != nil {
			logrus.Errorf("Get Link List Error: %s", err.Error())
			continue
		}
		for _, link := range links {
			if linkReadymap[link.Attrs().Name] {
				continue
			} else if topoInfo, ok := data.TopoInfo.LinkInfos[link.Attrs().Name]; ok {
				if topoInfo == nil {
					continue
				}
				err = SetInterfaceIP(link.Attrs().Name, topoInfo.V4Addr)
				if err != nil {
					logrus.Errorf("Set Addr %s to %s Error: %s", topoInfo.V4Addr, link.Attrs().Name, err.Error())
					continue
				}
				err = ReplaceDefaultRoute(link.Attrs().Index, topoInfo.V4Addr)
				if err != nil {
					logrus.Errorf("Set Default Gateway to %s Error: %s", topoInfo.V4Addr, err.Error())
					continue
				}
				linkReadymap[link.Attrs().Name] = true
			} else {
				linkReadymap[link.Attrs().Name] = false
			}
		}
		time.Sleep(checkGap)
	}
}
