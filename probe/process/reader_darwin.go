package process

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

const (
	lsofBinary    = "lsof"
	lsofFields    = "cn" // parseLSOF() depends on the order
	netstatBinary = "netstat"
)

// ProcReader is a /proc reader
type ProcReader struct {
	processes   []*Process
	connections []*Connection

	withProcs bool

	sync.RWMutex
}

// NewReader returns a Darwin (lsof-based) '/proc' reader
func NewReader(proc Dir, withProcs bool) *ProcReader {
	return &ProcReader{
		processes:   []*Process{},
		connections: []*Connection{},
		withProcs:   withProcs,
	}
}

// Tick reads the processes and connections
func (r *ProcReader) Read() error {
	newProcesses := []*Process{}
	newConnections := []*Connection{}

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

	newProcesses, err = parseLSOF(string(output))
	if err != nil {
		return err
	}

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
	newConnections = parseDarwinNetstat(string(out))

	if r.withProcs {
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
					newConnections[i].Process = proc
				}
			}
		}
	}

	r.Lock()
	defer r.Unlock()
	r.processes = newProcesses
	r.connections = newConnections

	return nil
}

// Tick updates the processes and connections lists
func (r *ProcReader) Tick() error {
	return r.Read()
}

// Close closes the Darwin "/proc" reader
func (ProcReader) Close() error {
	return nil
}

// Processes walks through the processes
func (ProcReader) Processes(f func(Process)) error {
	r.RLock()
	defer r.RUnlock()

	for _, p := range r.processes {
		f(*p)
	}
	return nil
}

// Connections walks through the connections
func (r *ProcReader) Connections(f func(Connection)) error {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.connections {
		f(*c)
	}
	return nil
}

func parseLSOF(output string) (map[string]*Process, error) {
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
			processes[addresses[0]] = &Process{
				PID:  process.PID,
				Comm: process.Comm,
			}

		default:
			return nil, fmt.Errorf("unexpected lsof field: %c in %#v", field, value)
		}
	}
	return processes, nil
}
