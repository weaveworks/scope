package host

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/report"
)

// Uname is swappable for mocking in tests.
var Uname = syscall.Uname

// GetKernelVersion returns the kernel version as reported by uname.
var GetKernelVersion = func() (string, error) {
	var utsname syscall.Utsname
	if err := Uname(&utsname); err != nil {
		return "unknown", err
	}
	return fmt.Sprintf("%s %s", charsToString(utsname.Release), charsToString(utsname.Version)), nil
}

// GetLoad returns the current load averages as metrics.
var GetLoad = func() report.Metrics {
	buf, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil
	}
	now := time.Now()
	toks := strings.Fields(string(buf))
	if len(toks) < 3 {
		return nil
	}
	one, err := strconv.ParseFloat(toks[0], 64)
	if err != nil {
		return nil
	}
	five, err := strconv.ParseFloat(toks[1], 64)
	if err != nil {
		return nil
	}
	fifteen, err := strconv.ParseFloat(toks[2], 64)
	if err != nil {
		return nil
	}
	return report.Metrics{
		Load1:  report.MakeMetric().Add(now, one),
		Load5:  report.MakeMetric().Add(now, five),
		Load15: report.MakeMetric().Add(now, fifteen),
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
