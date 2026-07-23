// systemstats.go
package boot

import (
	"bufio"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type SystemStats struct {
	Arch          string `json:"arch"`
	Distro        string `json:"distro"`
	Kernel        string `json:"kernel"`
	CPUUsage      int    `json:"cpuUsage"`
	MemoryUsage   int    `json:"memoryUsage"`
	DiskUsage     int    `json:"diskUsage"`
	Temperature   int    `json:"temperature"`
	NetworkStatus string `json:"networkStatus"`
	Timestamp     string `json:"timestamp"`
}

func GetStats() *SystemStats {
	stats := SystemStats{
		Arch:      getArch(),
		Distro:    getDistro(),
		Kernel:    getKernel(),
		Timestamp: getTimestamp(),
	}

	if cpu, err := getCPUUsage(); err != nil {
		stats.CPUUsage = 0
	} else {
		stats.CPUUsage = cpu
	}

	if mem, err := getMemoryUsage(); err != nil {
		stats.MemoryUsage = 0
	} else {
		stats.MemoryUsage = mem
	}

	if disk, err := getDiskUsage(); err != nil {
		stats.DiskUsage = 0
	} else {
		stats.DiskUsage = disk
	}

	if temp, err := getTemperature(); err != nil {
		stats.Temperature = 0
	} else {
		stats.Temperature = temp
	}

	if net, err := getNetworkStatus(); err != nil {
		stats.NetworkStatus = "inactive"
	} else {
		stats.NetworkStatus = net
	}

	return &stats
}

// REFACTORED INTERNALS: All functions use Session.RunCommand

func getArch() string {
	if out, err := runOnSession("uname -m"); err == nil && out != "" {
		return out
	}
	return runtime.GOARCH
}

func getCPUUsage() (int, error) {
	first, err := readProcStat()
	if err != nil {
		return 0, err
	}
	time.Sleep(250 * time.Millisecond)
	second, err := readProcStat()
	if err != nil {
		return 0, err
	}

	totalDelta := second.total - first.total
	idleDelta := second.idle - first.idle
	if totalDelta <= 0 {
		return 0, fmt.Errorf("invalid cpu sample")
	}

	return clampPercent(int((100 * (totalDelta - idleDelta)) / totalDelta)), nil
}

func getMemoryUsage() (int, error) {
	out, err := runOnSession("cat /proc/meminfo")
	if err != nil {
		return 0, err
	}

	values := map[string]int{}
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		values[strings.TrimSuffix(fields[0], ":")] = value
	}

	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total <= 0 {
		return 0, fmt.Errorf("missing memory totals")
	}

	return clampPercent(int((100 * (total - available)) / total)), nil
}

func getDiskUsage() (int, error) {
	out, err := runOnSession("df -P /")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("missing disk usage")
	}

	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("unexpected df output")
	}

	usage := strings.TrimSuffix(fields[4], "%")
	value, err := strconv.Atoi(usage)
	if err != nil {
		return 0, err
	}
	return clampPercent(value), nil
}

func getTemperature() (int, error) {
	temperatureFiles := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
	}

	for _, file := range temperatureFiles {
		if data, err := runOnSession("cat " + file); err == nil && len(data) > 0 {
			if temp, err := strconv.Atoi(strings.TrimSpace(data)); err == nil {
				return temp / 1000, nil
			}
		}
	}
	return 0, fmt.Errorf("no thermal zone found")
}

func getNetworkStatus() (string, error) {
	// Evaluates exit code directly in the active shell environment
	_, err := runOnSession("ip route get 1.1.1.1")
	if err == nil {
		return "active", nil
	}
	return "inactive", nil
}

func getTimestamp() string {
	return time.Now().Local().Format(time.RFC1123)
}

func getDistro() string {
	if out, err := runOnSession(". /etc/os-release 2>/dev/null && printf '%s' \"$PRETTY_NAME\""); err == nil && out != "" {
		return out
	}
	return "unknown"
}

func getKernel() string {
	if out, err := runOnSession("uname -r"); err == nil && out != "" {
		return out
	}
	return "unknown"
}

type cpuSample struct {
	total int
	idle  int
}

func readProcStat() (cpuSample, error) {
	out, err := runOnSession("cat /proc/stat")
	if err != nil {
		return cpuSample{}, err
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		return cpuSample{}, fmt.Errorf("empty /proc/stat output")
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, fmt.Errorf("unexpected /proc/stat output layout")
	}

	sample := cpuSample{}
	for i, field := range fields[1:] {
		value, err := strconv.Atoi(field)
		if err != nil {
			return cpuSample{}, err
		}
		sample.total += value
		if i == 3 || i == 4 { // Idle time slots
			sample.idle += value
		}
	}
	return sample, nil
}

// Optimized Helper replacing runWSL process-spawning
func runOnSession(cmdStr string) (string, error) {
	if Session == nil {
		return "", fmt.Errorf("active session instance not bound yet")
	}
	return Session.RunCommand(cmdStr)
}

func clampPercent(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
