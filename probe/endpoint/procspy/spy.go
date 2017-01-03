// Package procspy lists TCP connections, and optionally tries to find the
// owning processes. Works on Linux (via /proc) and Darwin (via `lsof -i` and
// `netstat`). You'll need root to use Processes().
package procspy

import (
	"net"
)

// Connection is a (TCP) connection. The Proc struct might not be filled in.
type Connection struct {
	Transport     string
	LocalAddress  net.IP
	LocalPort     uint16
	RemoteAddress net.IP
	RemotePort    uint16
	inode         uint64
	Proc
}

// Proc is a single process with PID and process name.
type Proc struct {
	PID            uint
	Name           string
	NetNamespaceID uint64
}

// ConnIter is returned by Connections().
type ConnIter interface {
	Next() *Connection
}

// ConnectionScanner scans the system for established (TCP) connections
type ConnectionScanner interface {
	// Connections returns all established (TCP) connections.  If processes is
	// false we'll just list all TCP connections, and there is no need to be root.
	// If processes is true it'll additionally try to lookup the process owning the
	// connection, filling in the Proc field. You will need to run this as root to
	// find all processes.
	Connections(processes bool) (ConnIter, error)
	// Stops the scanning
	Stop()
}
