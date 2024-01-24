package utils

import (
	"NodeDaemon/model"
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

func GetHostResourceInfo() (*model.HostResourceRaw, error) {
	var ret = new(model.HostResourceRaw)

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

func GetInstanceResourceInfo(containerID string) (*model.InstanceResouceRaw, error) {
	pathBase := fmt.Sprintf(
		"/sys/fs/cgroup/system.slice/docker-%s.scope/",
		containerID,
	)
	cpuPath := path.Join(pathBase, "cpu.stat")
	memPath := path.Join(pathBase, "memory.current")
	swapPath := path.Join(pathBase, "memory.swap.current")

	mem, err := readUint64(memPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Memory Usage Of Container %s Error: %s", containerID, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	swap, err := readUint64(swapPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Swap Usage Of Container %s Error: %s", containerID, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}
	cpuUsage, err := readCpuUsage(cpuPath)
	if err != nil {
		errMsg := fmt.Sprintf("Read Cpu Usage Of Container %s Error: %s", containerID, err.Error())
		logrus.Error(errMsg)
		return nil, err
	}

	return &model.InstanceResouceRaw{
		CPUBusy:     cpuUsage / 1000,
		MemByte:     mem,
		SwapMemByte: swap,
	}, nil
}

func GetInstanceLinkResourceInfo(pid int) (map[string]*model.LinkResourceRaw, error) {
	ret := make(map[string]*model.LinkResourceRaw)
	filePath := fmt.Sprintf("/proc/%d/net/dev", pid)
	devStat, err := net.IOCountersByFile(true, filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Get Instance Link Resource Error: %s", err.Error())
		logrus.Error(errMsg)
		return ret, err
	}
	for _, v := range devStat {
		ret[v.Name] = &model.LinkResourceRaw{
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
