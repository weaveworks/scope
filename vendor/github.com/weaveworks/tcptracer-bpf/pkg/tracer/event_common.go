package tracer

import (
	"net"
)

type EventType uint32

// These constants should be in sync with the equivalent definitions in the ebpf program.
const (
	EventConnect EventType = 1
	EventAccept            = 2
	EventClose             = 3
)

func (e EventType) String() string {
	switch e {
	case EventConnect:
		return "connect"
	case EventAccept:
		return "accept"
	case EventClose:
		return "close"
	default:
		return "unknown"
	}
}

// TcpV4 represents a TCP event (connect, accept or close) on IPv4
type TcpV4 struct {
	Timestamp uint64    // Monotonic timestamp
	CPU       uint64    // CPU index
	Type      EventType // connect, accept or close
	Pid       uint32    // Process ID, who triggered the event
	Comm      string    // The process command (as in /proc/$pid/comm)
	SAddr     net.IP    // Local IP address
	DAddr     net.IP    // Remote IP address
	SPort     uint16    // Local TCP port
	DPort     uint16    // Remote TCP port
	NetNS     uint32    // Network namespace ID (as in /proc/$pid/ns/net)
}

// TcpV6 represents a TCP event (connect, accept or close) on IPv6
type TcpV6 struct {
	Timestamp uint64    // Monotonic timestamp
	CPU       uint64    // CPU index
	Type      EventType // connect, accept or close
	Pid       uint32    // Process ID, who triggered the event
	Comm      string    // The process command (as in /proc/$pid/comm)
	SAddr     net.IP    // Local IP address
	DAddr     net.IP    // Remote IP address
	SPort     uint16    // Local TCP port
	DPort     uint16    // Remote TCP port
	NetNS     uint32    // Network namespace ID (as in /proc/$pid/ns/net)
}
