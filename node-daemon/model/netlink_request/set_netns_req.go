package netreq

type SetNetNsReq struct {
	NetLinkRequestBase
	NamespacePid int
}

func CreateSetNetNsReq(linkIndex, linkNamespaceFd, namespacePid int) SetNetNsReq {
	return SetNetNsReq{
		NetLinkRequestBase: NetLinkRequestBase{
			NamespaceFd: linkNamespaceFd,
			LinkIndex:   linkIndex,
			RequestType: SetLinkState,
		},
		NamespacePid: namespacePid,
	}
}
