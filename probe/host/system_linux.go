package host

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Uname is swappable for mocking in tests.
var Uname = syscall.Uname

func charsToString(ca [65]int8) string {
	s := make([]byte, len(ca))
	var lens int
	for ; lens < len(ca); lens++ {
		if ca[lens] == 0 {
			break
		}
		s[lens] = uint8(ca[lens])
	}
	return string(s[0:lens])
}

// GetKernelVersion returns the kernel version as reported by uname.
var GetKernelVersion = func() (string, error) {
	var utsname syscall.Utsname
	if err := Uname(&utsname); err != nil {
		return "unknown", err
	}
	return fmt.Sprintf("%s %s", charsToString(utsname.Release), charsToString(utsname.Version)), nil
}

// GetLoad returns the current load averages in standard form.
var GetLoad = func() string {
	buf, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return "unknown"
	}
	toks := strings.Fields(string(buf))
	if len(toks) < 3 {
		return "unknown"
	}
	one, err := strconv.ParseFloat(toks[0], 64)
	if err != nil {
		return "unknown"
	}
	five, err := strconv.ParseFloat(toks[1], 64)
	if err != nil {
		return "unknown"
	}
	fifteen, err := strconv.ParseFloat(toks[2], 64)
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%.2f %.2f %.2f", one, five, fifteen)
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
