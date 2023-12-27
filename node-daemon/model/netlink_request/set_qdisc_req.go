package netreq

import "github.com/vishvananda/netlink"

const (
	AddQdisc = iota
	ReplaceQdisc
	DelQdisc
)

type SetQdiscReq struct {
	NetLinkRequestBase
	OperationType int
	QdiscInfo     netlink.Qdisc
}

func CreateSetQdiscReq(linkIndex, linkNamespacePid, operationType int, linkName string, qdiscInfo netlink.Qdisc) SetQdiscReq {
	return SetQdiscReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:    linkIndex,
			RequestType:  SetQdisc,
			LinkName:     linkName,
		},
		QdiscInfo:     qdiscInfo,
		OperationType: operationType,
	}
}
