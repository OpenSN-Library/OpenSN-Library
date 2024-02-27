package utils

import "net"


func CreateV4InetMask(prefixLen int) []byte {
	mask32 := ^(1<<(32-prefixLen) - 1)
	return net.IPv4Mask(
		byte(mask32>>24&0xff),
		byte(mask32>>16&0xff),
		byte(mask32>>8&0xff),
		byte(mask32&0xff),
	)

}