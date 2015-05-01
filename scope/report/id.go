package report

import (
	"net"
	"strings"
)

// IDAddresser is used to get the IP address from an addressID. Or nil.
type IDAddresser func(string) net.IP

// AddressIPPort translates "scope;ip;port" to the IP address. These are used
// by Process topologies.
func AddressIPPort(id string) net.IP {
	parts := strings.SplitN(id, ScopeDelim, 3)
	if len(parts) != 3 {
		return nil // hmm
	}
	return net.ParseIP(parts[1])
}

// AddressIP translates "scope;ip" to the IP address. These are used by
// Network topologies.
func AddressIP(id string) net.IP {
	parts := strings.SplitN(id, ScopeDelim, 2)
	if len(parts) != 2 {
		return nil // hmm
	}
	return net.ParseIP(parts[1])
}
