package procspy

import (
	"net"
	"os/exec"
	"strconv"

	"github.com/weaveworks/scope/probe/process"
)

const (
	netstatBinary = "netstat"
	lsofBinary    = "lsof"
)

// NewConnectionScanner creates a new Darwin ConnectionScanner
func NewConnectionScanner(_ process.Walker, processes bool) ConnectionScanner {
	return &darwinScanner{processes}
}

// NewSyncConnectionScanner creates a new synchronous Darwin ConnectionScanner
func NewSyncConnectionScanner(_ process.Walker, processes bool) ConnectionScanner {
	return &darwinScanner{processes}
}

type darwinScanner struct {
	processes bool
}

// Connections returns all established (TCP) connections.
func (s *darwinScanner) Connections() (ConnIter, error) {
	out, err := exec.Command(
		netstatBinary,
		"-n", // no number resolving
		"-W", // Wide output
		// "-l", // full IPv6 addresses // What does this do?
		"-p", "tcp", // only TCP
	).CombinedOutput()
	if err != nil {
		// Log.Infof("lsof error: %s", err)
		return nil, err
	}
	connections := parseDarwinNetstat(string(out))

	if s.processes {
		out, err := exec.Command(
			lsofBinary,
			"-i",       // only Internet files
			"-n", "-P", // no number resolving
			"-w",             // no warnings
			"-F", lsofFields, // \n based output of only the fields we want.
		).CombinedOutput()
		if err != nil {
			return nil, err
		}

		procs, err := parseLSOF(string(out))
		if err != nil {
			return nil, err
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

	f := fixedConnIter(connections)
	return &f, nil
}

// Nothing to stop since there's nothing running in the background
func (s *darwinScanner) Stop() {}
