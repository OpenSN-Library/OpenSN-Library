package netreq

type SetNetNsReq struct {
	NetLinkRequestBase
	TargetNamespacePid int
}

func CreateSetNetNsReq(linkIndex, linkNamespacePid, namespacePid int) SetNetNsReq {
	return SetNetNsReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespacePid: linkNamespacePid,
			LinkIndex:   linkIndex,
			RequestType: SetLinkState,
		},
		TargetNamespacePid: namespacePid,
	}
}
