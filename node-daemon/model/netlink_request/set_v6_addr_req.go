package netreq

type SetV6AddrReq struct {
	NetLinkRequestBase
	V6Addr    string
	PrefixLen int
}

func CreateSetV6AddrReq(linkNamespacePid int, linkName string, v6Addr string) *SetV6AddrReq {
	return &SetV6AddrReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			RequestType:  SetV6Addr,
			LinkName:     linkName,
		},
		V6Addr:    v6Addr,
	}
}
