package ifconfig

import (
	"fmt"
	"ground/utils"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func ReplaceDefaultRoute(linkIndex int, v4Addr string) error {
	gatewayIP := net.ParseIP(strings.Split(v4Addr, "/")[0])
	if gatewayIP[15]%4 == 1 {
		gatewayIP[15] += 1
	} else if gatewayIP[15]%4 == 2 {
		gatewayIP[15] -= 1
	}
	return netlink.RouteReplace(
		&netlink.Route{
			LinkIndex: linkIndex,
			Dst: &net.IPNet{
				IP:   []byte{0, 0, 0, 0},
				Mask: []byte{0, 0, 0, 0},
			},
			Gw: gatewayIP,
		},
	)
}

func DelInterfaceIPs(ifName string) error {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		logrus.Errorf("Find Link By Name %s Error: %s", ifName, err.Error())
		return err
	}
	addrList, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		logrus.Errorf("Find Addr List By Link Name %s Error: %s", ifName, err.Error())
		return err
	}
	for _, addr := range addrList {
		netlink.AddrDel(link, &addr)
	}
	return nil
}

func SetInterfaceIP(ifName string, v4Addr string) error {
	addr := strings.Split(v4Addr, "/")
	if len(addr) < 2 {
		err := fmt.Errorf("invalid ipv4 addr %s", v4Addr)
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", ifName, err.Error())
		return err
	}
	ip := net.ParseIP(addr[0])
	prefixLen, err := strconv.Atoi(addr[1])
	if err != nil {
		err = fmt.Errorf("invalid ipv4 addr prefix length %s", err.Error())
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", ifName, err.Error())
		return err
	}
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		logrus.Errorf("Find Link By Name %s Error: %s. Retry...", ifName, err.Error())
		return err
	}
	err = netlink.LinkSetUp(link)
	if err != nil {
		return err
	}
	netlinkAddr := netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: utils.CreateV4InetMask(prefixLen),
		},
	}
	err = netlink.AddrAdd(link, &netlinkAddr)
	if err != nil {
		logrus.Errorf("Set Link %s IPv4 Addr error: %s", ifName, err.Error())
		time.Sleep(time.Second)
	}
	return nil
}
