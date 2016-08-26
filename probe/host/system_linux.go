package host

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"

	"github.com/weaveworks/scope/common/marshal"
	"github.com/weaveworks/scope/report"
)

const kb = 1024

// Uname is swappable for mocking in tests.
var Uname = syscall.Uname

// GetKernelVersion returns the kernel version as reported by uname.
var GetKernelVersion = func() (string, error) {
	var utsname syscall.Utsname
	if err := Uname(&utsname); err != nil {
		return "unknown", err
	}
	release := marshal.FromUtsname(utsname.Release)
	version := marshal.FromUtsname(utsname.Version)
	return fmt.Sprintf("%s %s", release, version), nil
}

// GetLoad returns the current load averages as metrics.
var GetLoad = func(now time.Time) report.Metrics {
	buf, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil
	}
	toks := strings.Fields(string(buf))
	if len(toks) < 3 {
		return nil
	}
	one, err := strconv.ParseFloat(toks[0], 64)
	if err != nil {
		return nil
	}
	return report.Metrics{
		Load1: report.MakeSingletonMetric(now, one),
	}
}

// GetUptime returns the uptime of the host.
var GetUptime = func() (time.Duration, error) {
	buf, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(buf))
	if len(fields) != 2 {
		return 0, fmt.Errorf("invalid format: %s", string(buf))
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(uptime) * time.Second, nil
}

var previousStat = linuxproc.CPUStat{}

// GetCPUUsagePercent returns the percent cpu usage and max (i.e. 100% or 0 if unavailable)
var GetCPUUsagePercent = func() (float64, float64) {
	stat, err := linuxproc.ReadStat(ProcStat)
	if err != nil {
		return 0.0, 0.0
	}

	// From http://stackoverflow.com/questions/23367857/accurate-calculation-of-cpu-usage-given-in-percentage-in-linux
	var (
		currentStat = stat.CPUStatAll
		prevIdle    = previousStat.Idle + previousStat.IOWait
		idle        = currentStat.Idle + currentStat.IOWait
		prevNonIdle = (previousStat.User + previousStat.Nice + previousStat.System +
			previousStat.IRQ + previousStat.SoftIRQ + previousStat.Steal)
		nonIdle = (currentStat.User + currentStat.Nice + currentStat.System +
			currentStat.IRQ + currentStat.SoftIRQ + currentStat.Steal)
		prevTotal = prevIdle + prevNonIdle
		total     = idle + nonIdle
		// differentiate: actual value minus the previous one
		totald = total - prevTotal
		idled  = idle - prevIdle
	)
	previousStat = currentStat
	return float64(totald-idled) * 100. / float64(totald), 100.
}

// GetMemoryUsageBytes returns the bytes memory usage and max
var GetMemoryUsageBytes = func() (float64, float64) {
	meminfo, err := linuxproc.ReadMemInfo(ProcMemInfo)
	if err != nil {
		return 0.0, 0.0
	}

	used := meminfo.MemTotal - meminfo.MemFree - meminfo.Buffers - meminfo.Cached
	return float64(used * kb), float64(meminfo.MemTotal * kb)
}
