package main

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func main() {
	raw,_ := netns.Get()
	list, _ := netlink.LinkList()
	for _, v := range list {
		fmt.Printf("NetNs FD is %d, Interface Name is %s, IfIndex is %d\n", raw, v.Attrs().Name, v.Attrs().Index)
	}
	test0, _ := netns.GetFromName("test_0")
	defer test0.Close()
	test1, _ := netns.GetFromName("test_1")
	defer test1.Close()
	err := netns.Set(test0)
	if err != nil {
		panic(err)
	}
	list, _ = netlink.LinkList()
	for _, v := range list {
		fmt.Printf("NetNs FD is %d, Interface Name is %s, IfIndex is %d\n", test0, v.Attrs().Name, v.Attrs().Index)
	}
	err = netns.Set(test1)
	if err != nil {
		panic(err)
	}
	list, _ = netlink.LinkList()
	for _, v := range list {
		fmt.Printf("NetNs FD is %d, Interface Name is %s, IfIndex is %d\n", test1, v.Attrs().Name, v.Attrs().Index)
	}

}
