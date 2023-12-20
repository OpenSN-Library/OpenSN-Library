package netreq

type SetV4AddrReq struct {
	NetLinkRequestBase
	V4Addr uint32
	PrefixLen int
}


func CreateSetV4AddrReq(linkIndex, linkNamespacePid int, v4Addr uint32, prefixLen int) SetV4AddrReq {
	return SetV4AddrReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex: linkIndex,
			RequestType: SetLinkState,
		},
		V4Addr: v4Addr,
		PrefixLen: prefixLen,
	}
}