package report

import (
	"hash"
	"hash/fnv"
	"net"
	"strings"
	"sync"

	"github.com/bluele/gcache"
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

	// Key added to nodes to prevent them being joined with conntracked connections
	DoesNotMakeConnections = "does_not_make_connections"
)

var (
	idCache = gcache.New(1024).LRU().Build()
	hashers = sync.Pool{
		New: func() interface{} {
			return fnv.New64a()
		},
	}
)

func lookupID(part1, part2, part3 string, f func() string) string {
	h := hashers.Get().(hash.Hash64)
	h.Write([]byte(part1))
	h.Write([]byte(part2))
	h.Write([]byte(part3))
	sum := h.Sum64()
	var result string
	if id, err := idCache.Get(sum); id != nil && err != nil {
		result = id.(string)
	} else {
		result = f()
		idCache.Set(sum, result)
	}
	h.Reset()
	hashers.Put(h)
	return result
}

// MakeEndpointNodeID produces an endpoint node ID from its composite parts.
func MakeEndpointNodeID(hostID, address, port string) string {
	return lookupID(hostID, address, port, func() string {
		return MakeAddressNodeID(hostID, address) + ScopeDelim + port
	})
}

// MakeAddressNodeID produces an address node ID from its composite parts.
func MakeAddressNodeID(hostID, address string) string {
	var scope string

	// Loopback addresses and addresses explicitly marked as
	// local get scoped by hostID
	addressIP := net.ParseIP(address)
	if addressIP != nil && LocalNetworks.Contains(addressIP) {
		scope = hostID
	} else if isLoopback(address) {
		scope = hostID
	}

	return scope + ScopeDelim + address
}

// MakeScopedEndpointNodeID is like MakeEndpointNodeID, but it always
// prefixes the ID witha scope.
func MakeScopedEndpointNodeID(hostID, address, port string) string {
	return hostID + ScopeDelim + address + ScopeDelim + port
}

// MakeScopedAddressNodeID is like MakeAddressNodeID, but it always
// prefixes the ID witha scope.
func MakeScopedAddressNodeID(hostID, address string) string {
	return hostID + ScopeDelim + address
}

// MakeProcessNodeID produces a process node ID from its composite parts.
func MakeProcessNodeID(hostID, pid string) string {
	return hostID + ScopeDelim + pid
}

var (
	// MakeHostNodeID produces a host node ID from its composite parts.
	MakeHostNodeID = makeSingleComponentID("host")

	// ParseHostNodeID parses a host node ID
	ParseHostNodeID = parseSingleComponentID("host")

	// MakeContainerNodeID produces a container node ID from its composite parts.
	MakeContainerNodeID = makeSingleComponentID("container")

	// ParseContainerNodeID parses a container node ID
	ParseContainerNodeID = parseSingleComponentID("container")

	// MakeContainerImageNodeID produces a container image node ID from its composite parts.
	MakeContainerImageNodeID = makeSingleComponentID("container_image")

	// ParseContainerImageNodeID parses a container image node ID
	ParseContainerImageNodeID = parseSingleComponentID("container_image")

	// MakePodNodeID produces a pod node ID from its composite parts.
	MakePodNodeID = makeSingleComponentID("pod")

	// ParsePodNodeID parses a pod node ID
	ParsePodNodeID = parseSingleComponentID("pod")

	// MakeServiceNodeID produces a service node ID from its composite parts.
	MakeServiceNodeID = makeSingleComponentID("service")

	// ParseServiceNodeID parses a service node ID
	ParseServiceNodeID = parseSingleComponentID("service")

	// MakeDeploymentNodeID produces a deployment node ID from its composite parts.
	MakeDeploymentNodeID = makeSingleComponentID("deployment")

	// ParseDeploymentNodeID parses a deployment node ID
	ParseDeploymentNodeID = parseSingleComponentID("deployment")

	// MakeReplicaSetNodeID produces a replica set node ID from its composite parts.
	MakeReplicaSetNodeID = makeSingleComponentID("replica_set")

	// ParseReplicaSetNodeID parses a replica set node ID
	ParseReplicaSetNodeID = parseSingleComponentID("replica_set")
)

// makeSingleComponentID makes a single-component node id encoder
func makeSingleComponentID(tag string) func(string) string {
	return func(id string) string {
		return id + ScopeDelim + "<" + tag + ">"
	}
}

// parseSingleComponentID makes a single-component node id decoder
func parseSingleComponentID(tag string) func(string) (string, bool) {
	return func(id string) (string, bool) {
		fields := strings.SplitN(id, ScopeDelim, 2)
		if len(fields) != 2 || fields[1] != "<"+tag+">" {
			return "", false
		}
		return fields[0], true
	}
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

// ParseEndpointNodeID produces the host ID, address, and port and remainder
// (typically an address) from an endpoint node ID. Note that hostID may be
// blank.
func ParseEndpointNodeID(endpointNodeID string) (hostID, address, port string, ok bool) {
	fields := strings.SplitN(endpointNodeID, ScopeDelim, 3)
	if len(fields) != 3 {
		return "", "", "", false
	}
	return fields[0], fields[1], fields[2], true
}

// ParseAddressNodeID produces the host ID, address from an address node ID.
func ParseAddressNodeID(addressNodeID string) (hostID, address string, ok bool) {
	fields := strings.SplitN(addressNodeID, ScopeDelim, 2)
	if len(fields) != 2 {
		return "", "", false
	}
	return fields[0], fields[1], true
}

// ExtractHostID extracts the host id from Node
func ExtractHostID(m Node) string {
	hostNodeID, _ := m.Latest.Lookup(HostNodeID)
	hostID, _, _ := ParseNodeID(hostNodeID)
	return hostID
}

func isLoopback(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.IsLoopback()
}
