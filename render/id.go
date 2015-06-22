package render

import (
	"fmt"
	"strings"
)

// MakeEndpointID makes an endpoint node ID for rendered nodes.
func MakeEndpointID(hostID, addr, port string) string {
	return fmt.Sprintf("endpoint:%s:%s:%s", hostID, addr, port)
}

// MakeProcessID makes a process node ID for rendered nodes.
func MakeProcessID(hostID, pid string) string {
	return fmt.Sprintf("process:%s:%s", hostID, pid)
}

// MakeAddressID makes an address node ID for rendered nodes.
func MakeAddressID(hostID, addr string) string {
	return fmt.Sprintf("address:%s:%s", hostID, addr)
}

// MakeHostID makes a host node ID for rendered nodes.
func MakeHostID(hostID string) string {
	return fmt.Sprintf("host:%s", hostID)
}

// MakePseudoNodeID produces a pseudo node ID from its composite parts,
// for use in rendered nodes.
func MakePseudoNodeID(parts ...string) string {
	return strings.Join(append([]string{"pseudo"}, parts...), ":")
}
