package utils

import (
	"sync"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type Interface struct {
	Name string
	Index int
	netNsFD int
	fdLock *sync.Mutex
	
}

func (i *Interface) GetNetNs() netns.NsHandle {
	i.fdLock.Lock()
	return netns.NsHandle(i.netNsFD)
}

func (i *Interface) Unlock() {
	i.fdLock.Unlock()
}

func GetAllInterface(containerIDs []string) ([]Interface,error){
	rawNs,err := netns.Get()
	var ret []Interface
	if err != nil {
		return nil,err
	}
	if rawNs.IsOpen() {
		defer rawNs.Close()
	}

	ifs,err := netlink.LinkList()

	if err != nil {
		return nil,err
	}

	for _,v := range ifs {
		ret = append(ret, Interface{
			Name: v.Attrs().Name,
			Index: v.Attrs().Index,
			netNsFD: int(rawNs),
		})
	}

	for _,id := range containerIDs {

		ns,err := netns.GetFromDocker(id)
		if err != nil {
			return nil,err
		}
		if ns.IsOpen() {
			defer ns.Close()
		}

		err = netns.Set(ns)

		if err != nil {
			return nil,err
		}

		ifs,err = netlink.LinkList()

		if err != nil {
			return nil,err
		}

		for _,v := range ifs {
			ret = append(ret, Interface{
				Name: v.Attrs().Name,
				Index: v.Attrs().Index,
				netNsFD: int(ns),
			})
		}

		if err != nil {
			return nil,err
		}
	}
	
	return ret,netns.Set(rawNs)
}