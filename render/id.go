package render

import (
	"strings"

	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	IncomingInternetID = "in-theinternet"
	OutgoingInternetID = "out-theinternet"
)

// IsInternetNode determines whether the node represents the Internet.
func IsInternetNode(n report.Node) bool {
	return n.ID == IncomingInternetID || n.ID == OutgoingInternetID
}

// MakePseudoNodeID joins the parts of an id into the id of a pseudonode
func MakePseudoNodeID(parts ...string) string {
	return strings.Join(append([]string{"pseudo"}, parts...), ":")
}

// ParsePseudoNodeID returns the joined id parts of a pseudonode
// ID. If the ID is not recognisable as a pseudonode ID, it is
// returned as is, with the returned bool set to false. That is
// convenient because not all pseudonode IDs actually follow the
// format produced by MakePseudoNodeID.
func ParsePseudoNodeID(nodeID string) (string, bool) {
	// Not using strings.SplitN() to avoid a heap allocation
	pos := strings.Index(nodeID, ":")
	if pos == -1 || nodeID[:pos] != "pseudo" {
		return nodeID, false
	}
	return nodeID[pos+1:], true
}

// MakeGroupNodeTopology joins the parts of a group topology into the topology of a group node
func MakeGroupNodeTopology(originalTopology, key string) string {
	return strings.Join([]string{"group", originalTopology, key}, ":")
}

// ParseGroupNodeTopology returns the parts of a group topology.
func ParseGroupNodeTopology(topology string) (string, string, bool) {
	parts := strings.Split(topology, ":")
	if len(parts) != 3 || parts[0] != "group" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// NewDerivedNode makes a node based on node, but with a new ID
func NewDerivedNode(id string, node report.Node) report.Node {
	return report.MakeNode(id).WithChildren(node.ChildIDs.Add(node.ID))
}

// NewDerivedPseudoNode makes a new pseudo node with the node as a child
func NewDerivedPseudoNode(id string, node report.Node) report.Node {
	output := NewDerivedNode(id, node).WithTopology(Pseudo)
	return output
}

func pseudoNodeID(rpt report.Report, n report.Node, local report.Networks) (string, bool) {
	_, addr, _, ok := report.ParseEndpointNodeID(n.ID)
	if !ok {
		return "", false
	}

	if id, ok := externalNodeID(rpt, n, addr, local); ok {
		return id, ok
	}

	// due to https://github.com/weaveworks/scope/issues/1323 we are dropping
	// all non-external pseudo nodes for now.
	return "", false
}

// figure out if a node should be considered external and returns an ID which can be used to create a pseudo node
func externalNodeID(rpt report.Report, n report.Node, addr string, local report.Networks) (string, bool) {
	// First, check if it's a known service and emit a a specific node if it
	// is. This needs to be done before checking IPs since known services can
	// live in the same network, see https://github.com/weaveworks/scope/issues/2163
	if hostname, found := rpt.DNS.FirstMatch(n.ID, isKnownService); found {
		return ServiceNodeIDPrefix + hostname, true
	}

	// If the dstNodeAddr is not in a network local to this report, we emit an
	// internet pseudoNode
	// Create a buffer on the stack of this function, so we don't need to allocate in ParseIP
	var into [5]byte // one extra byte to save a memory allocation in critbitgo
	if ip := report.ParseIP([]byte(addr), into[:4]); ip != nil && !local.Contains(ip) {
		// emit one internet node for incoming, one for outgoing
		if len(n.Adjacency) > 0 {
			return IncomingInternetID, true
		}
		return OutgoingInternetID, true
	}

	// The node is not external
	return "", false
}
