package netreq

import (
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

const (
	SetLinkState = iota
	SetV4Addr
	SetV6Addr
	SetNetNs
	SetQdisc
)

type NetLinkRequest interface {
	GetLinkNamespaceFd() int
	GetLinkIndex() int
	GetRequestType() int
	ReleseFd() error
}

type NetLinkRequestBase struct {
	NamespaceFd int
	LinkIndex int
	RequestType int
}

func (r *NetLinkRequestBase)GetLinkNamespaceFd() int {
	return r.NamespaceFd
}

func (r *NetLinkRequestBase)GetLinkIndex() int {
	return r.LinkIndex
}

func (r *NetLinkRequestBase)GetRequestType() int {
	return r.RequestType
}

func (r *NetLinkRequestBase)ReleseFd() error {
	nsHandle := netns.NsHandle(r.NamespaceFd)
	if nsHandle.IsOpen() {
		err := nsHandle.Close()
		if err != nil {
			logrus.Errorf("Close Net Namespace Fd %d Error: %s",r.NamespaceFd,err.Error())
			return err
		}
	}
	return nil
}