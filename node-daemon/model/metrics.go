package model

type HostResourceRaw struct {
	CPUTotal    float64
	CPUBusy     float64
	MemByte     uint64
	SwapMemByte uint64
}

type InstanceResouceRaw struct {
	CPUBusy     float64
	MemByte     uint64
	SwapMemByte uint64
}

type LinkResourceRaw struct {
	RecvByte     uint64
	SendByte     uint64
	RecvPack     uint64
	SendPack     uint64
	RecvErrPack  uint64
	SendErrPack  uint64
	RecvDropPack uint64
	SendDropPack uint64
}

type HostResource struct {
	CPUUsage    float64
	MemByte     uint64
	SwapMemByte uint64
}

type InstanceResouce struct {
	CPUUsage    float64
	MemByte     uint64
	SwapMemByte uint64
}

type LinkResource struct {
	RecvBps     float64
	SendBps     float64
	RecvPps     float64
	SendPps     float64
	RecvErrPps  float64
	SendErrPps  float64
	RecvDropPps float64
	SendDropPps float64
}
