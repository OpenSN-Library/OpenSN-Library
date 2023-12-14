package bpf

import "github.com/vishvananda/netlink"

const (
	CLSACT_PARENT = 0xfffffff1
 	CLSACT_HANDLE = 0xffff0000
)


type Clsact struct {
	netlink.QdiscAttrs
}

func (qdisc *Clsact) Attrs() *netlink.QdiscAttrs {
	return &qdisc.QdiscAttrs
}

func (qdisc *Clsact) Type() string {
	return "clsact"
}

func CreateClsactQdisc(ifName string) (*Clsact, error) {
	link,err := netlink.LinkByName("docker0")
	if err != nil {
		return nil,err
	}
	ret := new(Clsact)

	ret.LinkIndex = link.Attrs().Index
	ret.Handle = CLSACT_HANDLE
	ret.Parent = CLSACT_PARENT

	return ret,nil
}