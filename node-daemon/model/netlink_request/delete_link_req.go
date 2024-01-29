package netreq

type DeleteLinkReq struct {
	NetLinkRequestBase
}

func CreateDeleteLinkReq(linkNamespacePid int,linkName string) DeleteLinkReq {
	return DeleteLinkReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			RequestType: DeleteLink,
			LinkName: linkName,
		},
	}
}