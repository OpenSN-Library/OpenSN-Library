package utils

import "net"

func CreateV4Inet(ip uint32, prefixLen int) *net.IPNet {
	mask32 := ^(1<<(32-prefixLen) - 1)
	return &net.IPNet{
		IP: net.IPv4(
			byte(ip>>24&0xff),
			byte(ip>>16&0xff),
			byte(ip>>8&0xff),
			byte(ip&0xff),
		),
		Mask: net.IPv4Mask(
			byte(mask32>>24&0xff),
			byte(mask32>>16&0xff),
			byte(mask32>>8&0xff),
			byte(mask32&0xff),
		),
	}
}

func CreateV6Inet(ip uint64, prefixLen int) *net.IPNet {
	mask64 := ^(uint64(1)<<(64-prefixLen) - 1)
	return &net.IPNet{
		IP: []byte{
			byte(ip >> 56 & 0xff),
			byte(ip >> 48 & 0xff),
			byte(ip >> 40 & 0xff),
			byte(ip >> 32 & 0xff),
			byte(ip >> 24 & 0xff),
			byte(ip >> 16 & 0xff),
			byte(ip >> 8 & 0xff),
			byte(ip & 0xff),
		},
		Mask: []byte{
			byte(mask64 >> 56 & 0xff),
			byte(mask64 >> 48 & 0xff),
			byte(mask64 >> 40 & 0xff),
			byte(mask64 >> 32 & 0xff),
			byte(mask64 >> 24 & 0xff),
			byte(mask64 >> 16 & 0xff),
			byte(mask64 >> 8 & 0xff),
			byte(mask64 & 0xff),
		},
	}
}
