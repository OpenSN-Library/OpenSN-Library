package bpf_test

import (
	"redirector/bpf"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestAddClsact(t *testing.T) {
	
	clsact,err := bpf.CreateClsactQdisc("docker0")

	if err != nil {
		t.Error(err)
	}


	err = netlink.QdiscAdd(clsact)

	if err != nil {
		t.Error(err)
	}

	netlink.QdiscDel(clsact)

	if err != nil {
		t.Error(err)
	}
}

func TestLoadBpf2Qdisc(t *testing.T) {
	err := bpf.LoadBpf2Qdisc("ens36",bpf.EGRESS)
	if err != nil {
		t.Error(err)
	}
}