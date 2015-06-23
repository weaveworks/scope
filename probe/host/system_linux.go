package host

import (
	"fmt"
	"strconv"
	"strings"
)

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
