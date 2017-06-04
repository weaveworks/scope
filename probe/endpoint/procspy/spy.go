// Package procspy lists TCP connections, and optionally tries to find the
// owning processes. Works on Linux (via /proc) and Darwin (via `lsof -i` and
// `netstat`). You'll need root to use Processes().
package procspy

import (
	"net"
)

const (
	// according to /include/net/tcp_states.h
	tcpEstablished = 1
	tcpFinWait1    = 4
	tcpFinWait2    = 5
	tcpCloseWait   = 8
)

// Connection is a (TCP) connection. The Proc struct might not be filled in.
type Connection struct {
	Transport     string
	LocalAddress  net.IP
	LocalPort     uint16
	RemoteAddress net.IP
	RemotePort    uint16
	Inode         uint64
	Proc          Proc
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
	// Connections returns all established (TCP) connections.
	Connections() (ConnIter, error)
	// Stops the scanning
	Stop()
}
