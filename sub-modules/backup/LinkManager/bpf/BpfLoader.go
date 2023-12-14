package bpf

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

//go:embed bpf.o
var _BPF_OBJECT_BYTES []byte

type BPF_OBJECT_TYPE struct {
	RedirectMap *ebpf.Map	`ebpf:"acc_map"`
	IngressProgram   *ebpf.Program	`ebpf:"tc_ingress"`
	EgressProgram	 *ebpf.Program	`ebpf:"tc_egress"`
}

var BPF_OBJECT BPF_OBJECT_TYPE
const BPF_PROGRAM_NAME = "bpf-redirector"

const (
	INGRESS = 0
	EGRESS  = 1
)

func init() {
	reader := bytes.NewReader(_BPF_OBJECT_BYTES)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		errMsg := fmt.Sprintf("Load BPF Object from file error %s", err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}

	err = spec.LoadAndAssign(&BPF_OBJECT, nil)

	if err != nil {
		errMsg := fmt.Sprintf("Assign BPF Object to struct error %s", err.Error())
		logrus.Error(errMsg)
		panic(errMsg)
	}

}

func AddClsactQdisc(ifName string) error {
	clsact,err := CreateClsactQdisc(ifName)

	if err != nil {
		return err
	}

	return netlink.QdiscAdd(clsact)
}

func DelClsactQdisc(ifName string) error {
	clsact,err := CreateClsactQdisc(ifName)

	if err != nil {
		return err
	}

	return netlink.QdiscDel(clsact)
}

func LoadBpf2Qdisc(ifName string, hookDirection int) error {
	
	link,err := netlink.LinkByName(ifName)

	if err != nil {
		return err
	}

	filter := &netlink.BpfFilter{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:   0,
			Handle:    1,
			Protocol:  unix.ETH_P_ALL,
			Priority:  1,
		},
		Name:         fmt.Sprintf("%s-%s", BPF_PROGRAM_NAME, link.Attrs().Name),
		DirectAction: true,
	}

	if hookDirection == INGRESS {
		filter.Fd = BPF_OBJECT.IngressProgram.FD()
	} else if hookDirection == EGRESS {
		filter.Fd = BPF_OBJECT.EgressProgram.FD()
	} else {
		return errors.New("invalid hook direction")
	}

	if err := netlink.FilterAdd(filter); err != nil {
		return fmt.Errorf("adding tc filter: %w", err)
	}
	return nil
}

func UnloadBpfFromQdisc(ifName string, hookDirection int) error {
	link,err := netlink.LinkByName(ifName)

	if err != nil {
		return err
	}

	filter := &netlink.BpfFilter{
		FilterAttrs: netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:   0,
			Handle:    1,
			Protocol:  unix.ETH_P_ALL,
			Priority:  1,
		},
		Name:         fmt.Sprintf("%s-%s", BPF_PROGRAM_NAME, link.Attrs().Name),
		DirectAction: true,
	}

	if hookDirection == INGRESS {
		filter.Fd = BPF_OBJECT.IngressProgram.FD()
	} else if hookDirection == EGRESS {
		filter.Fd = BPF_OBJECT.EgressProgram.FD()
	} else {
		return errors.New("invalid hook direction")
	}

	if err := netlink.FilterDel(filter); err != nil {
		return fmt.Errorf("deleting tc filter: %w", err)
	}
	return nil
}