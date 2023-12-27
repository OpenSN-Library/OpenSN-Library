package netreq

const (
	SetLinkState = iota
	SetV4Addr
	SetV6Addr
	SetNetNs
	SetQdisc
	DeleteLink
)

type NetLinkRequest interface {
	GetLinkNamespacePid() int
	GetLinkIndex() int
	GetLinkName() string
	GetRequestType() int
}

type NetLinkRequestBase struct {
	NamespacePid int
	LinkIndex    int
	LinkName     string
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

func (r *NetLinkRequestBase) GetLinkName() string {
	return r.LinkName
}
