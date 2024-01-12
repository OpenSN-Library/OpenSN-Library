package utils

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var NumberRegex = regexp.MustCompile(`^([1-9][0-9]*)([KMGTkmgt]?)$`)

func CreateV4InetMask(prefixLen int) []byte {
	mask32 := ^(1<<(32-prefixLen) - 1)
	return net.IPv4Mask(
		byte(mask32>>24&0xff),
		byte(mask32>>16&0xff),
		byte(mask32>>8&0xff),
		byte(mask32&0xff),
	)

}

func CreateV6InetMask(prefixLen int) []byte {
	mask64 := ^(uint64(1)<<(64-prefixLen) - 1)
	return []byte{
		byte(mask64 >> 56 & 0xff),
		byte(mask64 >> 48 & 0xff),
		byte(mask64 >> 40 & 0xff),
		byte(mask64 >> 32 & 0xff),
		byte(mask64 >> 24 & 0xff),
		byte(mask64 >> 16 & 0xff),
		byte(mask64 >> 8 & 0xff),
		byte(mask64 & 0xff),
	}
}

func FormartIPAddr(addr []byte, prefixLen int) (string, error) {
	if len(addr) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d/%d", addr[0], addr[1], addr[2], addr[3], prefixLen), nil
	} else if len(addr) == 16 {
		return fmt.Sprintf(
			"%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x:%x%x/%d",
			addr[0], addr[1],
			addr[2], addr[3],
			addr[4], addr[5],
			addr[6], addr[7],
			addr[8], addr[9],
			addr[10], addr[11],
			addr[12], addr[13],
			addr[14], addr[15],
			prefixLen,
		), nil
	} else {
		return "", fmt.Errorf("invalid length %d for ip addr", len(addr))
	}
}

func ByteArrayAdd(origin []byte, delta uint32) []byte {
	array := make([]byte, len(origin))
	copy(array, origin)
	for i := len(array) - 1; i >= 0; i-- {
		sum := uint32(array[i]) + delta
		array[i] = byte(sum % 256)
		delta = uint32(sum) / 256
		if delta == 0 {
			break
		}
	}
	return array
}

func ParseBinNumber(s string) (int64, error) {
	parts := NumberRegex.FindStringSubmatch(s)
	if len(parts) <= 0 {
		return 0, fmt.Errorf("invalid number string %s, support [1-9][0-9]*[KMGTkmgt]?", s)
	}
	var conffienct int64 = 1

	if len(parts) > 2 {
		switch strings.ToLower(parts[2]) {
		case "k":
			conffienct <<= 10
		case "m":
			conffienct <<= 20
		case "g":
			conffienct <<= 30
		case "t":
			conffienct <<= 40
		}
	}

	base, _ := strconv.ParseInt(parts[1], 10, 64)
	return conffienct * base, nil
}

func ParseDecNumber(s string) (int64, error) {
	parts := NumberRegex.FindStringSubmatch(s)
	if len(parts) <= 0 {
		return 0, fmt.Errorf("invalid number string %s, support [1-9][0-9]*[KMGTkmgt]?", s)
	}
	var conffienct int64 = 1

	if len(parts) > 2 {
		switch strings.ToLower(parts[2]) {
		case "k":
			conffienct *= 1e3
		case "m":
			conffienct *= 1e6
		case "g":
			conffienct *= 1e9
		case "t":
			conffienct *= 1e12
		}
	}

	base, _ := strconv.ParseInt(parts[1], 10, 64)
	return conffienct * base, nil
}

func FormatIPv4(addr []byte) string {
	if len(addr) < 4 {
		return ""
	}
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		addr[0],
		addr[1],
		addr[2],
		addr[3],
	)
}

func FormatIPv6(addr []byte) string {
	if len(addr) < 16 {
		return ""
	}
	ret := fmt.Sprintf("%02x%02x", addr[0], addr[1])
	for i := 2; i < 16; i += 2 {
		ret += fmt.Sprintf(":%02x%02x", addr[i], addr[i+1])
	}
	return ret
}

func FormatMacAddr(addr []byte) string {
	if len(addr) < 6 {
		return ""
	}
	return fmt.Sprintf(
		"%02x:%02x:%02x:%02x:%02x:%02x",
		addr[0],
		addr[1],
		addr[2],
		addr[3],
		addr[4],
		addr[5],
	)
}
