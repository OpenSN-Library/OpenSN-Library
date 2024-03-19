package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

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

func readCpuUsage(filePath string) (float64, error) {
	found := false
	var usageUsec float64
	f, err := os.Open(filePath)
	if err != nil {
		return usageUsec, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\n")
		split := strings.Split(line, " ")
		if len(split) < 2 {
			continue
		}
		if split[0] == "usage_usec" {
			usageUsec, err = strconv.ParseFloat(split[1], 64)
			if err != nil {
				return usageUsec, err
			}
			found = true
			break
		}
	}

	if !found {
		return usageUsec, fmt.Errorf("field usage_usec not found")
	}

	return usageUsec, nil
}

func readUint64(filePath string) (uint64, error) {
	var ret uint64
	f, err := os.Open(filePath)
	if err != nil {
		return ret, err
	}
	defer f.Close()

	seq, err := io.ReadAll(f)
	uint64Str := strings.TrimRight(string(seq), "\n")
	if err != nil {
		return ret, err
	}
	return strconv.ParseUint(uint64Str, 10, 64)
}

func GetHostResourceInfo() (*HostResourceRaw, error) {
	var ret = new(HostResourceRaw)

	cpuStat, err := cpu.Times(false)

	if err != nil {
		errMsg := fmt.Sprintf("Get Host CPU Times Info Error: %s", err.Error())
		logrus.Error(errMsg)
		return ret, err
	}

	ret.CPUTotal = cpuStat[0].Total() - cpuStat[0].Guest - cpuStat[0].GuestNice
	ret.CPUBusy = ret.CPUTotal - cpuStat[0].Idle - cpuStat[0].Iowait
	memStat, err := mem.VirtualMemory()
	if err != nil {
		errMsg := fmt.Sprintf("Get Host Memory Times Info Error: %s", err.Error())
		logrus.Error(errMsg)
		return ret, nil
	}
	ret.MemByte = memStat.Total - memStat.Free
	ret.SwapMemByte = memStat.SwapTotal - memStat.SwapFree
	ret.CPUBusy *= 1000
	ret.CPUTotal *= 1000
	return ret, nil
}

func GetInstanceResourceInfo(instanceID string, pid int) (*InstanceResouceRaw, error) {
	cgroupProc, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pid))
	defer cgroupProc.Close()
	if err != nil {
		errMsg := fmt.Sprintf("Open Cgroup Mount Dir Record Of %s Error: %s", instanceID, err.Error())
		logrus.Warn(errMsg)
		return nil, err
	}
	bytes, err := io.ReadAll(cgroupProc)

	if err != nil {
		errMsg := fmt.Sprintf("Read Cgroup Mount Dir Of %s Error: %s", instanceID, err.Error())
		logrus.Warn(errMsg)
		return nil, err
	}

	splitedStr := strings.Split(string(bytes), "::")
	if len(splitedStr) < 2 {
		errMsg := fmt.Sprintf("Read Cgroup Mount Dir Of %s Error: invalid format: %s", instanceID, string(bytes))
		logrus.Warn(errMsg)
		return nil, err
	}

	pathBase := fmt.Sprintf(
		"/sys/fs/cgroup/%s/",
		splitedStr[1],
	)
	cpuPath := path.Join(pathBase, "cpu.stat")
	memPath := path.Join(pathBase, "memory.current")
	swapPath := path.Join(pathBase, "memory.swap.current")

	mem, err := readUint64(memPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Memory Usage Of %s Error: %s", instanceID, err.Error())
		logrus.Warn(errMsg)
		return nil, err
	}
	swap, err := readUint64(swapPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Swap Usage Of %s Error: %s", instanceID, err.Error())
		logrus.Warn(errMsg)
		return nil, err
	}
	cpuUsage, err := readCpuUsage(cpuPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Cpu Usage Of %s Error: %s", instanceID, err.Error())
		logrus.Warn(errMsg)
		return nil, err
	}

	return &InstanceResouceRaw{
		CPUBusy:     cpuUsage / 1000,
		MemByte:     mem,
		SwapMemByte: swap,
	}, nil
}

func GetInstanceLinkResourceInfo(pid int) (map[string]*LinkResourceRaw, error) {
	ret := make(map[string]*LinkResourceRaw)
	filePath := fmt.Sprintf("/proc/%d/net/dev", pid)
	devStat, err := net.IOCountersByFile(true, filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Get Instance Link Resource Error: %s", err.Error())
		logrus.Warn(errMsg)
		return ret, err
	}
	for _, v := range devStat {
		ret[v.Name] = &LinkResourceRaw{
			RecvByte:     v.BytesRecv,
			SendByte:     v.BytesSent,
			RecvPack:     v.PacketsRecv,
			SendPack:     v.PacketsSent,
			SendErrPack:  v.Errout,
			RecvErrPack:  v.Errin,
			RecvDropPack: v.Dropin,
			SendDropPack: v.Dropout,
		}
	}
	return ret, nil
}
