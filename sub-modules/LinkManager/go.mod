module LinkManager

go 1.20

replace dependencies/netlink v1.1.0 => github.com/vishvananda/netlink v1.1.0

require (
	github.com/cilium/ebpf v0.12.3
	github.com/sirupsen/logrus v1.9.3
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	golang.org/x/sys v0.14.1-0.20231108175955-e4099bfacb8c
)

require (
	github.com/stretchr/testify v1.8.1 // indirect
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // indirect
)
