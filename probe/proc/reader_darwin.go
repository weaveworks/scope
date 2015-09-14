package proc

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type reader struct{}

// NewReader returns a Darwin (lsof-based) '/proc' reader
func NewReader(proc Dir) Reader {
	return &reader{}
}

const (
	lsofBinary    = "lsof"
	lsofFields    = "cn" // parseLSOF() depends on the order
	netstatBinary = "netstat"
	lsofBinary    = "lsof"
)

func (reader) Processes(f func(Process)) error {
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
		f(process)
	}
	return nil
}

func (r *reader) Connections(withProcs bool, f func(Connection)) error {
	out, err := exec.Command(
		netstatBinary,
		"-n", // no number resolving
		"-W", // Wide output
		// "-l", // full IPv6 addresses // What does this do?
		"-p", "tcp", // only TCP
	).CombinedOutput()
	if err != nil {
		return err
	}
	connections := parseDarwinNetstat(string(out))

	if withProcs {
		out, err := exec.Command(
			lsofBinary,
			"-i",       // only Internet files
			"-n", "-P", // no number resolving
			"-w",             // no warnings
			"-F", lsofFields, // \n based output of only the fields we want.
		).CombinedOutput()
		if err != nil {
			return err
		}

		procs, err := parseLSOF(string(out))
		if err != nil {
			return err
		}
		for local, proc := range procs {
			for i, c := range connections {
				localAddr := net.JoinHostPort(
					c.LocalAddress.String(),
					strconv.Itoa(int(c.LocalPort)),
				)
				if localAddr == local {
					connections[i].Proc = proc
				}
			}
		}
	}

	for _, c := range connections {
		f(c)
	}
	return nil
}

// Close closes the Darwin "/proc" reader
func (reader) Close() error {
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
			process.Comm = value

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
				Comm: process.Comm,
			}

		default:
			return nil, fmt.Errorf("unexpected lsof field: %c in %#v", field, value)
		}
	}
	return processes, nil
}
