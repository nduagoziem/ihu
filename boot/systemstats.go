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

func GetStats(session *BootWSL) *SystemStats {
	stats := SystemStats{
		Arch:      getArch(session),
		Distro:    getDistro(session),
		Kernel:    getKernel(session),
		Timestamp: getTimestamp(),
	}

	if cpu, err := getCPUUsage(session); err != nil {
		stats.CPUUsage = 0
	} else {
		stats.CPUUsage = cpu
	}

	if mem, err := getMemoryUsage(session); err != nil {
		stats.MemoryUsage = 0
	} else {
		stats.MemoryUsage = mem
	}

	if disk, err := getDiskUsage(session); err != nil {
		stats.DiskUsage = 0
	} else {
		stats.DiskUsage = disk
	}

	if temp, err := getTemperature(session); err != nil {
		stats.Temperature = 0
	} else {
		stats.Temperature = temp
	}

	if net, err := getNetworkStatus(session); err != nil {
		stats.NetworkStatus = "inactive"
	} else {
		stats.NetworkStatus = net
	}

	return &stats
}

// REFACTORED INTERNALS: All functions use the app-owned session.

func getArch(session *BootWSL) string {
	if out, err := runOnSession(session, "uname -m"); err == nil && out != "" {
		return out
	}
	return runtime.GOARCH
}

func getCPUUsage(session *BootWSL) (int, error) {
	first, err := readProcStat(session)
	if err != nil {
		return 0, err
	}
	time.Sleep(250 * time.Millisecond)
	second, err := readProcStat(session)
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

func getMemoryUsage(session *BootWSL) (int, error) {
	out, err := runOnSession(session, "cat /proc/meminfo")
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

func getDiskUsage(session *BootWSL) (int, error) {
	out, err := runOnSession(session, "df -P /")
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

func getTemperature(session *BootWSL) (int, error) {
	temperatureFiles := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
	}

	for _, file := range temperatureFiles {
		if data, err := runOnSession(session, "cat "+file); err == nil && len(data) > 0 {
			if temp, err := strconv.Atoi(strings.TrimSpace(data)); err == nil {
				return temp / 1000, nil
			}
		}
	}
	return 0, fmt.Errorf("no thermal zone found")
}

func getNetworkStatus(session *BootWSL) (string, error) {
	// Evaluates exit code directly in the active shell environment
	_, err := runOnSession(session, "ip route get 1.1.1.1")
	if err == nil {
		return "active", nil
	}
	return "inactive", nil
}

func getTimestamp() string {
	return time.Now().Local().Format(time.RFC1123)
}

func getDistro(session *BootWSL) string {
	if out, err := runOnSession(session, ". /etc/os-release 2>/dev/null && printf '%s' \"$PRETTY_NAME\""); err == nil && out != "" {
		return out
	}
	return "unknown"
}

func getKernel(session *BootWSL) string {
	if out, err := runOnSession(session, "uname -r"); err == nil && out != "" {
		return out
	}
	return "unknown"
}

type cpuSample struct {
	total int
	idle  int
}

func readProcStat(session *BootWSL) (cpuSample, error) {
	out, err := runOnSession(session, "cat /proc/stat")
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
func runOnSession(session *BootWSL, cmdStr string) (string, error) {
	if session == nil {
		return "", fmt.Errorf("active session instance not bound yet")
	}
	return session.RunCommand(cmdStr)
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
