package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// NewWalker returns a Darwin (lsof-based) walker.
func NewWalker(_ string, _ bool) Walker {
	return &walker{}
}

type walker struct{}

const (
	lsofBinary    = "lsof"
	lsofFields    = "cn" // parseLSOF() depends on the order
	netstatBinary = "netstat"
)

// These functions copied from procspy.

// IsProcInAccept returns true if the process has a at least one thread
// blocked on the accept() system call
func IsProcInAccept(procRoot, pid string) (ret bool) {
	// Not implemented on darwin
	return false
}

func (walker) Walk(f func(Process, Process)) error {
	output, err := exec.Command(
		lsofBinary,
		"-i",       // only Internet files
		"-n", "-P", // no number resolving
		"-w",             // no warnings
		"-F", lsofFields, // \n based output of only the fields we want.
	).CombinedOutput()
	if err != nil {
		return err
	}

	processes, err := parseLSOF(string(output))
	if err != nil {
		return err
	}

	for _, process := range processes {
		f(process, Process{})
	}
	return nil
}

func parseLSOF(output string) (map[string]Process, error) {
	var (
		processes = map[string]Process{} // Local addr -> Proc
		process   Process
	)
	for _, line := range strings.Split(output, "\n") {
		if len(line) <= 1 {
			continue
		}

		var (
			field = line[0]
			value = line[1:]
		)
		switch field {
		case 'p':
			pid, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid 'p' field in lsof output: %#v", value)
			}
			process.PID = pid

		case 'c':
			process.Name = value

		case 'n':
			// 'n' is the last field, with '-F cn'
			// format examples:
			// "192.168.2.111:44013->54.229.241.196:80"
			// "[2003:45:2b57:8900:1869:2947:f942:aba7]:55711->[2a00:1450:4008:c01::11]:443"
			// "*:111" <- a listen
			addresses := strings.SplitN(value, "->", 2)
			if len(addresses) != 2 {
				// That's a listen entry.
				continue
			}
			processes[addresses[0]] = Process{
				PID:  process.PID,
				Name: process.Name,
			}

		default:
			return nil, fmt.Errorf("unexpected lsof field: %c in %#v", field, value)
		}
	}
	return processes, nil
}

// GetDeltaTotalJiffies returns 0 - darwin doesn't have jiffies.
func GetDeltaTotalJiffies() (uint64, float64, error) {
	return 0, 0.0, nil
}
