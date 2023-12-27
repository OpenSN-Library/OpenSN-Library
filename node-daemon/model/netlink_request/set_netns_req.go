package netreq

type SetNetNsReq struct {
	NetLinkRequestBase
	TargetNamespacePid int
}

func CreateSetNetNsReq(linkIndex, linkNamespacePid, namespacePid int, linkName string) SetNetNsReq {
	return SetNetNsReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:    linkIndex,
			RequestType:  SetNetNs,
			LinkName:     linkName,
		},
		TargetNamespacePid: namespacePid,
	}
}
