// Package procspy lists TCP connections, and optionally tries to find the
// owning processes. Works on Linux (via /proc) and Darwin (via `lsof -i` and
// `netstat`). You'll need root to use Processes().
package procspy

import (
	"net"
)

const (
	// TCPEstablished according to /include/net/tcp_states.h
	TCPEstablished = 1
)

// Connection is a (TCP) connection. The Proc struct might not be filled in.
type Connection struct {
	Transport     string
	LocalAddress  net.IP
	LocalPort     uint16
	RemoteAddress net.IP
	RemotePort    uint16
	Inode         uint64
	Proc
}

// Proc is a single process with PID and process name.
type Proc struct {
	PID  uint
	Name string
}

// ConnIter is returned by Connections().
type ConnIter interface {
	Next() *Connection
}

// Connections returns all established (TCP) connections.  If processes is
// false we'll just list all TCP connections, and there is no need to be root.
// If processes is true it'll additionally try to lookup the process owning the
// connection, filling in the Proc field. You will need to run this as root to
// find all processes.
func Connections(processes bool) (ConnIter, error) {
	return cbConnections(processes)
}
