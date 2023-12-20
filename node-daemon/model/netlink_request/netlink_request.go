package netreq

const (
	SetLinkState = iota
	SetV4Addr
	SetV6Addr
	SetNetNs
	SetQdisc
)

type NetLinkRequest interface {
	GetLinkNamespacePid() int
	GetLinkIndex() int
	GetRequestType() int
}

type NetLinkRequestBase struct {
	NamespacePid int
	LinkIndex    int
	RequestType  int
}

func (r *NetLinkRequestBase) GetLinkNamespacePid() int {
	return r.NamespacePid
}

func (r *NetLinkRequestBase) GetLinkIndex() int {
	return r.LinkIndex
}

func (r *NetLinkRequestBase) GetRequestType() int {
	return r.RequestType
}
