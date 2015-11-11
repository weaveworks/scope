package host

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/weaveworks/scope/report"
)

var (
	unameRe  = regexp.MustCompile(`Darwin Kernel Version ([0-9\.]+)\:`)
	loadRe   = regexp.MustCompile(`load averages: ([0-9\.]+) ([0-9\.]+) ([0-9\.]+)`)
	uptimeRe = regexp.MustCompile(`up ([0-9]+) day[s]*,[ ]+([0-9]+)\:([0-9][0-9])`)
)

// GetKernelVersion returns the kernel version as reported by uname.
var GetKernelVersion = func() (string, error) {
	out, err := exec.Command("uname", "-v").CombinedOutput()
	if err != nil {
		return "Darwin unknown", err
	}
	matches := unameRe.FindAllStringSubmatch(string(out), -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 1 {
		return "Darwin unknown", nil
	}
	return fmt.Sprintf("Darwin %s", matches[0][1]), nil
}

// GetLoad returns the current load averages as metrics.
var GetLoad = func() report.Metrics {
	out, err := exec.Command("w").CombinedOutput()
	if err != nil {
		return nil
	}
	now := time.Now()
	matches := loadRe.FindAllStringSubmatch(string(out), -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 4 {
		return nil
	}

	one, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil {
		return nil
	}
	five, err := strconv.ParseFloat(matches[0][2], 64)
	if err != nil {
		return nil
	}
	fifteen, err := strconv.ParseFloat(matches[0][3], 64)
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
