package host

import (
	"fmt"
	"os/exec"
	"regexp"
)

var loadRe = regexp.MustCompile(`load average\: ([0-9\.]+), ([0-9\.]+), ([0-9\.]+)`)

func getKernelVersion() (string, error) {
	return "", nil
}

func getLoad() string {
	out, err := exec.Command("w").CombinedOutput()
	if err != nil {
		return "unknown"
	}
	matches := loadRe.FindAllStringSubmatch(string(out), -1)
	if matches == nil || len(matches) < 1 || len(matches[0]) < 4 {
		return "unknown"
	}
	return fmt.Sprintf("%s %s %s", matches[0][1], matches[0][2], matches[0][3])
}
