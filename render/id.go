package render

import (
	"net"
	"sort"
	"strings"

	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	TheInternetID      = "theinternet"
	IncomingInternetID = "in-" + TheInternetID
	OutgoingInternetID = "out-" + TheInternetID
)

// MakePseudoNodeID joins the parts of an id into the id of a pseudonode
func MakePseudoNodeID(parts ...string) string {
	return strings.Join(append([]string{"pseudo"}, parts...), ":")
}

// MakeGroupNodeTopology joins the parts of a group topology into the topology of a group node
func MakeGroupNodeTopology(originalTopology, key string) string {
	return strings.Join([]string{"group", originalTopology, key}, ":")
}

// NewDerivedNode makes a node based on node, but with a new ID
func NewDerivedNode(id string, node report.Node) report.Node {
	return report.MakeNode(id).WithChildren(node.Children.Add(node))
}

// NewDerivedPseudoNode makes a new pseudo node with the node as a child
func NewDerivedPseudoNode(id string, node report.Node) report.Node {
	output := NewDerivedNode(id, node).WithTopology(Pseudo)
	return output
}

// NewDerivedExternalNode figures out if a node should be considered external and creates the corresponding pseudo node
func NewDerivedExternalNode(n report.Node, addr string, local report.Networks) (report.Node, bool) {
	// First, check if it's a known service and emit a a specific node if it
	// is. This needs to be done before checking IPs since known services can
	// live in the same network, see https://github.com/weaveworks/scope/issues/2163
	for _, hostname := range DNSNames(n) {
		if isKnownService(hostname) {
			return NewDerivedPseudoNode(ServiceNodeIDPrefix+hostname, n), true
		}
	}

	// If the dstNodeAddr is not in a network local to this report, we emit an
	// internet pseudoNode
	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		// emit one internet node for incoming, one for outgoing
		if len(n.Adjacency) > 0 {
			return NewDerivedPseudoNode(IncomingInternetID, n), true
		}
		return NewDerivedPseudoNode(OutgoingInternetID, n), true
	}

	// The node is not external
	return report.Node{}, false
}

// DNSNames returns a prioritized list of snooped and reverse-resolved
// DNS names associated with node n.
func DNSNames(n report.Node) []string {
	snoopedNames, _ := n.Sets.Lookup(endpoint.SnoopedDNSNames)
	reverseNames, _ := n.Sets.Lookup(endpoint.ReverseDNSNames)
	// sort the names, to make selection for display more
	// deterministic
	sort.StringSlice(snoopedNames).Sort()
	sort.StringSlice(reverseNames).Sort()
	// prioritize snooped names
	return append(snoopedNames, reverseNames...)
}
