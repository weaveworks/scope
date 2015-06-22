package report

import (
	"net"
	"strings"
)

// TheInternet is used as a node ID to indicate a remote IP.
const TheInternet = "theinternet"

// Delimiters are used to separate parts of node IDs, to guarantee uniqueness
// in particular contexts.
const (
	// ScopeDelim is a general-purpose delimiter used within node IDs to
	// separate different contextual scopes. Different topologies have
	// different key structures.
	ScopeDelim = ";"

	// EdgeDelim separates two node IDs when they need to exist in the same key.
	// Concretely, it separates node IDs in keys that represent edges.
	EdgeDelim = "|"
)

// MakeAdjacencyID produces an adjacency ID from a node id.
func MakeAdjacencyID(srcNodeID string) string {
	return ">" + srcNodeID
}

// ParseAdjacencyID produces a node ID from an adjancency ID.
func ParseAdjacencyID(adjacencyID string) (string, bool) {
	if !strings.HasPrefix(adjacencyID, ">") {
		return "", false
	}
	return adjacencyID[1:], true
}

// MakeEdgeID produces an edge ID from composite parts.
func MakeEdgeID(srcNodeID, dstNodeID string) string {
	return srcNodeID + EdgeDelim + dstNodeID
}

// ParseEdgeID splits an edge ID to its composite parts.
func ParseEdgeID(edgeID string) (srcNodeID, dstNodeID string, ok bool) {
	fields := strings.SplitN(edgeID, EdgeDelim, 2)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// MakeEndpointNodeID produces an endpoint node ID from its composite parts.
func MakeEndpointNodeID(hostID, address, port string) string {
	return MakeAddressNodeID(hostID, address) + ScopeDelim + port
}

// MakeAddressNodeID produces an address node ID from its composite parts.
func MakeAddressNodeID(hostID, address string) string {
	var scope string

	// Loopback addresses and addresses explicity marked as
	// local get scoped by hostID
	addressIP := net.ParseIP(address)
	if addressIP != nil && LocalNetworks.Contains(addressIP) {
		scope = hostID
	} else if isLoopback(address) {
		scope = hostID
	}

	return scope + ScopeDelim + address
}

// MakeProcessNodeID produces a process node ID from its composite parts.
func MakeProcessNodeID(hostID, pid string) string {
	return hostID + ScopeDelim + pid
}

// MakeHostNodeID produces a host node ID from its composite parts.
func MakeHostNodeID(hostID string) string {
	// hostIDs come from the probe and are presumed to be globally-unique.
	// But, suffix something to elicit failures if we try to use probe host
	// IDs directly as node IDs in the host topology.
	return hostID + ScopeDelim + "<host>"
}

// MakeContainerNodeID produces a container node ID from its composite parts.
func MakeContainerNodeID(hostID, containerID string) string {
	return hostID + ScopeDelim + containerID
}

// MakeOverlayNodeID produces an overlay topology node ID from a router peer's
// name, which is assumed to be globally unique.
func MakeOverlayNodeID(peerName string) string {
	return "#" + peerName
}

// ParseNodeID produces the host ID and remainder (typically an address) from
// a node ID. Note that hostID may be blank.
func ParseNodeID(nodeID string) (hostID string, remainder string, ok bool) {
	fields := strings.SplitN(nodeID, ScopeDelim, 2)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// ExtractHostID extracts the host id from NodeMetadata
func ExtractHostID(m NodeMetadata) string {
	hostid, _, _ := ParseNodeID(m[HostNodeID])
	return hostid
}

// IDAddresser tries to convert a node ID to a net.IP, if possible.
type IDAddresser func(string) net.IP

// EndpointIDAddresser converts an endpoint node ID to an IP.
func EndpointIDAddresser(id string) net.IP {
	fields := strings.SplitN(id, ScopeDelim, 3)
	if len(fields) != 3 {
		//log.Printf("EndpointIDAddresser: bad input %q", id)
		return nil
	}
	return net.ParseIP(fields[1])
}

// AddressIDAddresser converts an address node ID to an IP.
func AddressIDAddresser(id string) net.IP {
	fields := strings.SplitN(id, ScopeDelim, 2)
	if len(fields) != 2 {
		//log.Printf("AddressIDAddresser: bad input %q", id)
		return nil
	}
	return net.ParseIP(fields[1])
}

func isLoopback(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.IsLoopback()
}
