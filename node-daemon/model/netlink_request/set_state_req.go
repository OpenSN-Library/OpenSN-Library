package netreq

type SetStateReq struct {
	NetLinkRequestBase
	Enable bool
}

func CreateSetStateReq(linkIndex, linkNamespacePid int, state bool) SetStateReq {
	return SetStateReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex: linkIndex,
			RequestType: SetLinkState,
		},
		Enable: state,
	}
}