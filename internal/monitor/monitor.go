package monitor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Snapshot struct {
	CPU       float64
	MemUsed   uint64
	MemTotal  uint64
	SwapUsed  uint64
	SwapTotal uint64
	DiskUsed  uint64
	DiskTotal uint64
	Load      string
	Uptime    time.Duration
}

func Sample() (Snapshot, error) {
	var s Snapshot
	a, err := readCPU()
	if err != nil {
		return s, err
	}
	time.Sleep(200 * time.Millisecond)
	b, err := readCPU()
	if err != nil {
		return s, err
	}
	s.CPU = cpuDelta(a, b)
	s.MemUsed, s.MemTotal, s.SwapUsed, s.SwapTotal = readMem()
	s.DiskUsed, s.DiskTotal = readDisk("/")
	s.Load = readLoad()
	s.Uptime = readUptime()
	return s, nil
}

type cpuTimes struct{ idle, total uint64 }

func readCPU() (cpuTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}
		var t cpuTimes
		for i, v := range fields[1:] {
			n, _ := strconv.ParseUint(v, 10, 64)
			t.total += n
			if i == 3 || i == 4 {
				t.idle += n
			}
		}
		return t, nil
	}
	return cpuTimes{}, fmt.Errorf("cpu stats not found")
}

func cpuDelta(a, b cpuTimes) float64 {
	dt := float64(b.total - a.total)
	di := float64(b.idle - a.idle)
	if dt <= 0 {
		return 0
	}
	return (1 - di/dt) * 100
}

func readMem() (memUsed, memTotal, swapUsed, swapTotal uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer f.Close()
	m := map[string]uint64{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		n, _ := strconv.ParseUint(fields[1], 10, 64)
		m[strings.TrimSuffix(fields[0], ":")] = n * 1024
	}
	memTotal = m["MemTotal"]
	memUsed = memTotal - m["MemAvailable"]
	swapTotal = m["SwapTotal"]
	swapUsed = swapTotal - m["SwapFree"]
	return
}

func readDisk(path string) (used, total uint64) {
	var st syscall.Statfs_t
	if syscall.Statfs(path, &st) != nil {
		return
	}
	total = st.Blocks * uint64(st.Bsize)
	used = total - st.Bavail*uint64(st.Bsize)
	return
}

func readLoad() string {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "?"
	}
	fields := strings.Fields(string(raw))
	if len(fields) < 3 {
		return "?"
	}
	return strings.Join(fields[:3], " ")
}

func readUptime() time.Duration {
	raw, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return time.Duration(secs) * time.Second
}
