package netreq

type SetV6AddrReq struct {
	NetLinkRequestBase
	V6Addr uint64
	PrefixLen int
}

func CreateSetV6AddrReq(linkIndex, linkNamespaceFd int, v6Addr uint64, prefixLen int) SetV6AddrReq {
	return SetV6AddrReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespaceFd: linkNamespaceFd,
			LinkIndex: linkIndex,
			RequestType: SetLinkState,
		},
		V6Addr: v6Addr,
		PrefixLen: prefixLen,
	}
}