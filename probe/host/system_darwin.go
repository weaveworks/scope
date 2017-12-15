package host

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/weaveworks/scope/report"
)

var (
	loadRe   = regexp.MustCompile(`load averages: ([0-9\.]+) ([0-9\.]+) ([0-9\.]+)`)
	uptimeRe = regexp.MustCompile(`up ([0-9]+) day[s]*,[ ]+([0-9]+)\:([0-9][0-9])`)
)

// GetKernelReleaseAndVersion returns the kernel version as reported by uname.
var GetKernelReleaseAndVersion = func() (string, string, error) {
	release, err := exec.Command("uname", "-r").CombinedOutput()
	if err != nil {
		return "unknown", "unknown", err
	}
	release = bytes.Trim(release, " \n")
	version, err := exec.Command("uname", "-v").CombinedOutput()
	if err != nil {
		return string(release), "unknown", err
	}
	version = bytes.Trim(version, " \n")
	return string(release), string(version), nil
}

// GetPrettyName extracts the PRETTY_NAME from /etc/os-release
var GetPrettyName = func () (string, error) {

	return "unknown", nil
}

// GetLoad returns the current load averages as metrics.
var GetLoad = func(now time.Time) report.Metrics {
	out, err := exec.Command("w").CombinedOutput()
	if err != nil {
		return nil
	}
	matches := loadRe.FindAllStringSubmatch(string(out), -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 4 {
		return nil
	}

	one, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil {
		return nil
	}
	return report.Metrics{
		Load1: report.MakeSingletonMetric(now, one),
	}
}

// GetUptime returns the uptime of the host.
var GetUptime = func() (time.Duration, error) {
	out, err := exec.Command("w").CombinedOutput()
	if err != nil {
		return 0, err
	}
	matches := uptimeRe.FindAllStringSubmatch(string(out), -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 4 {
		return 0, err
	}
	d, err := strconv.Atoi(matches[0][1])
	if err != nil {
		return 0, err
	}
	h, err := strconv.Atoi(matches[0][2])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(matches[0][3])
	if err != nil {
		return 0, err
	}
	return (time.Duration(d) * 24 * time.Hour) + (time.Duration(h) * time.Hour) + (time.Duration(m) * time.Minute), nil
}

// GetCPUUsagePercent returns the percent cpu usage and max (i.e. 100% or 0 if unavailable)
var GetCPUUsagePercent = func() (float64, float64) {
	return 0.0, 0.0
}

// GetMemoryUsageBytes returns the bytes memory usage and max
var GetMemoryUsageBytes = func() (float64, float64) {
	return 0.0, 0.0
}
