package netreq

type DeleteLinkReq struct {
	NetLinkRequestBase
}

func CreateDeleteLinkReq(linkIndex, linkNamespacePid int) DeleteLinkReq {
	return DeleteLinkReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:   linkIndex,
			RequestType: SetLinkState,
		},
	}
}