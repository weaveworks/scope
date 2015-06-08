package report

import (
	"fmt"
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

// MakeAdjacencyID produces an adjacency ID from composite parts.
func MakeAdjacencyID(hostID, srcNodeID string) string {
	// Here we rely on the fact that every possible source node ID has the
	// host ID as the first scope-delimited field, and therefore don't
	// duplicate that information.
	return ">" + srcNodeID
}

// ParseAdjacencyID splits an adjacency ID to its composite parts.
func ParseAdjacencyID(adjacencyID string) (hostID, srcNodeID string, ok bool) {
	if !strings.HasPrefix(adjacencyID, ">") {
		return "", "", false
	}
	// This relies on every node ID having hostID as its first scoped field.
	adjacencyID = adjacencyID[1:]
	fields := strings.SplitN(adjacencyID, ScopeDelim, 2)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], adjacencyID, true
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
	return hostID + ScopeDelim + address
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

// MakePseudoNodeID produces a pseudo node ID from its composite parts.
func MakePseudoNodeID(parts ...string) string {
	return strings.Join(append([]string{"pseudo"}, parts...), ScopeDelim)
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

// PanicIDAddresser will panic if it's ever called. It's used in topologies
// where there are never any edges, and so it's nonsensical to try and extract
// IPs from the node IDs.
func PanicIDAddresser(id string) net.IP {
	panic(fmt.Sprintf("PanicIDAddresser called on %q", id))
}
