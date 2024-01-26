package arranger

// import (
// 	"NodeDaemon/model"
// 	"NodeDaemon/utils"

// 	"github.com/sirupsen/logrus"
// )

// var (
// 	v4AddrStart = []byte{10, 0, 0, 0}
// 	v6AddrStart = []byte{0xfe, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
// )

// func ArrangeV4Addr(namespace *model.Namespace, prefixLen int) {
// 	var delta uint32 = (1 << (32 - prefixLen))
// 	prev := v4AddrStart
// 	for i := range namespace.LinkConfig {
// 		for j := 0; j < 2; j++ {
// 			addr, _ := utils.FormartIPAddr(utils.ByteArrayAdd(prev, uint32(j+1)), 30)
// 			namespace.LinkConfig[i].AddressInfos[j].V4Addr = addr
// 			logrus.Infof("Set IPv4 Addr %s to Link %s", addr, namespace.LinkConfig[i].LinkID)
// 		}
// 		prev = utils.ByteArrayAdd(prev, delta)
// 	}
// }

// func ArrangeV6Addr(namespace *model.Namespace, prefixLen int) {
// 	var delta uint32 = (1 << (64 - prefixLen))
// 	prev := v6AddrStart
// 	for i := range namespace.LinkConfig {
// 		for j := 0; j < 2; j++ {
// 			addr, _ := utils.FormartIPAddr(utils.ByteArrayAdd(prev, uint32(j+1)), 62)
// 			namespace.LinkConfig[i].AddressInfos[j].V6Addr = addr
// 			logrus.Infof("Set IPv6 Addr %s to Link %s", addr, namespace.LinkConfig[i].LinkID)
// 		}
// 		prev = utils.ByteArrayAdd(prev, delta)
// 	}
// }
