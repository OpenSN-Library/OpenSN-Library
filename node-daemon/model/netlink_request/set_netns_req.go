package netreq

type SetNetNsReq struct {
	NetLinkRequestBase
	TargetNamespacePid int
}

func CreateSetNetNsReq(linkNamespacePid, namespacePid int, linkName string) *SetNetNsReq {
	return &SetNetNsReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			RequestType:  SetNetNs,
			LinkName:     linkName,
		},
		TargetNamespacePid: namespacePid,
	}
}
