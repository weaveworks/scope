package proc

import (
	"fmt"
	"net"
)

const (
	tcpEstablished = 1 // according to /include/net/tcp_states.h
)

// Connection is a (TCP) connection. The 'Process' struct might not be filled in.
type Connection struct {
	Transport     string
	LocalAddress  net.IP
	LocalPort     uint16
	RemoteAddress net.IP
	RemotePort    uint16
	inode         uint64
	Process
}

// Copy returns a copy of a connection
func (c Connection) Copy() Connection {
	dupIP := func(ip net.IP) net.IP {
		dup := make(net.IP, len(ip))
		copy(dup, ip)
		return dup
	}

	c.LocalAddress = dupIP(c.LocalAddress)
	c.RemoteAddress = dupIP(c.RemoteAddress)
	return c
}

// String returns the string repr
func (c Connection) String() string {
	return fmt.Sprintf("%s:%d - %s:%d %s#%d",
		c.LocalAddress, c.LocalPort,
		c.RemoteAddress, c.RemotePort,
		c.Transport, c.inode)
}
