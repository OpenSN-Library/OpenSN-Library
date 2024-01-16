package netreq

type SetV4AddrReq struct {
	NetLinkRequestBase
	V4Addr    string
	PrefixLen int
}

func CreateSetV4AddrReq(linkIndex, linkNamespacePid int, linkName string, v4Addr string) *SetV4AddrReq {
	return &SetV4AddrReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:    linkIndex,
			RequestType:  SetV4Addr,
			LinkName:     linkName,
		},
		V4Addr:    v4Addr,
	}
}
