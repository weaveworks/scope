package host

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
)

// Uname exported for testing
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

func getKernelVersion() (string, error) {
	var utsname syscall.Utsname
	if err := Uname(&utsname); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", charsToString(utsname.Release), charsToString(utsname.Version)), nil
}

func getLoad() string {
	buf, err := ReadFile(ProcLoad)
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
