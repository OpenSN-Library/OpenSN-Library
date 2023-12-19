package netreq

type SetStateReq struct {
	NetLinkRequestBase
	Enable bool
}

func CreateSetStateReq(linkIndex, linkNamespaceFd int, state bool) SetStateReq {
	return SetStateReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespaceFd: linkNamespaceFd,
			LinkIndex: linkIndex,
			RequestType: SetLinkState,
		},
		Enable: state,
	}
}