package netreq

type SetStateReq struct {
	NetLinkRequestBase
	Enable bool
}

func CreateSetStateReq(linkNamespacePid int, linkName string, state bool) *SetStateReq {
	return &SetStateReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			RequestType:  SetLinkState,
			LinkName:     linkName,
		},
		Enable: state,
	}
}
