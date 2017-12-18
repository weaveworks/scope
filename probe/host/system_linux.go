package host

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"regexp"

	linuxproc "github.com/c9s/goprocinfo/linux"

	"github.com/weaveworks/scope/report"

	"golang.org/x/sys/unix"
)

const kb = 1024

// Uname is swappable for mocking in tests.
var Uname = unix.Uname

// GetKernelReleaseAndVersion returns the kernel version as reported by uname.
var GetKernelReleaseAndVersion = func() (string, string, error) {
	var utsname unix.Utsname
	if err := Uname(&utsname); err != nil {
		return "unknown", "unknown", err
	}
	release := utsname.Release[:bytes.IndexByte(utsname.Release[:], 0)]
	version := utsname.Version[:bytes.IndexByte(utsname.Version[:], 0)]
	return string(release), string(version), nil
}

// GetPrettyName extracts the PRETTY_NAME from host's /etc/os-release
// SPEC: https://www.freedesktop.org/software/systemd/man/os-release.html
var GetPrettyName = func () (string, error) {

	var buffer bytes.Buffer

	buf, err := ioutil.ReadFile("/var/run/scope/host-os-release")
	if err != nil {
		return "unknown", nil
	}

	// If can't find host-os-release, get container's os-release
	if buf == nil {
		bufContainer, err := ioutil.ReadFile("/etc/os-release")
		if err != nil {
			return "unknown", nil
		}
		buf = bufContainer
	}

	// Pretty Name Fallback Flow:
	// 1. $PRETTY_NAME
	// 2. $NAME $VERSION
	// 3. $ID $VERSION_ID
	// 4. runtime.GOOS

	prettyNameParse := regexp.MustCompile("PRETTY_NAME=(.+?)\n")
	nameParse := regexp.MustCompile("NAME=(.+?)\n")
	versionParse := regexp.MustCompile("VERSION=(.+?)\n")
	IDParse := regexp.MustCompile("ID=(.+?)\n")
	versionIDParse := regexp.MustCompile("VERSION_ID=(.+?)\n")

	prettyName := prettyNameParse.FindStringSubmatch(string(buf))
	name := nameParse.FindStringSubmatch(string(buf))
	version := versionParse.FindStringSubmatch(string(buf))
	prettyID := IDParse.FindStringSubmatch(string(buf))
	versionID := versionIDParse.FindStringSubmatch(string(buf))

	if prettyName == nil {
		if name == nil || version == nil {
			if prettyID == nil || versionID == nil {
				return string(buf), nil
			} else {
				buffer.WriteString(string(prettyID[1]))
				buffer.WriteString(" ")
				buffer.WriteString(string(versionID[1]))
			}
		} else {
			buffer.WriteString(string(name[1]))
			buffer.WriteString(" ")
			buffer.WriteString(string(version[1]))
		}
	} else {
		buffer.WriteString(string(prettyName[1]))
	}

	// Remove quotation marks from return value
	quoteParse := regexp.MustCompile("\"")
	return quoteParse.ReplaceAllString(buffer.String(), ""), nil
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
