package netreq

type DeleteLinkReq struct {
	NetLinkRequestBase
}

func CreateDeleteLinkReq(linkIndex, linkNamespacePid int,linkName string) DeleteLinkReq {
	return DeleteLinkReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:   linkIndex,
			RequestType: DeleteLink,
			LinkName: linkName,
		},
	}
}