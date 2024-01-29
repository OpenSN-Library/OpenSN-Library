package netreq

import "github.com/vishvananda/netlink"

type SetQdiscReq struct {
	NetLinkRequestBase
	OperationType int
	QdiscInfo     netlink.Qdisc
}

func CreateSetQdiscReq(linkNamespacePid int, linkName string, qdiscInfo netlink.Qdisc) *SetQdiscReq {
	return &SetQdiscReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			RequestType:  SetQdisc,
			LinkName:     linkName,
		},
		QdiscInfo:     qdiscInfo,
	}
}
