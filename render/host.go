package render

import (
	"github.com/weaveworks/scope/report"
)

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology.
var HostRenderer = MakeReduce(
	MakeMap(
		MapEndpoint2Host,
		EndpointRenderer,
	),
	MakeMap(
		MapX2Host,
		ProcessRenderer,
	),
	MakeMap(
		MapX2Host,
		ContainerRenderer,
	),
	MakeMap(
		MapX2Host,
		ContainerImageRenderer,
	),
	MakeMap(
		MapX2Host,
		PodRenderer,
	),
	SelectHost,
)

// MapX2Host maps any Nodes to host Nodes.
//
// If this function is given a node without a hostname
// (including other pseudo nodes), it will drop the node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapX2Host(n report.Node, _ report.Networks) report.Nodes {
	// Don't propagate all pseudo nodes - we do this in MapEndpoint2Host
	if n.Topology == Pseudo {
		return report.Nodes{}
	}
	ids, _ := n.Parents.Lookup(report.Host)
	results := report.Nodes{}
	for _, id := range ids {
		result := NewDerivedNode(id, n).
			WithTopology(report.Host).
			WithSet(report.HostNodeIDs, report.MakeStringSet(id)).
			WithCounters(map[string]int{n.Topology: 1})
		result.Children = report.MakeNodeSet(n)
		results[id] = result
	}
	return results
}

// MapEndpoint2Host takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
func MapEndpoint2Host(n report.Node, local report.Networks) report.Nodes {
	// Nodes without a host are treated as pseudo nodes
	_, ok := n.Parents.Lookup(report.Host)
	if !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	return MapX2Host(n, local)
}
