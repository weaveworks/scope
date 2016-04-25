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
		ColorConnected(ProcessRenderer),
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
	hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID)
	if !ok {
		return report.Nodes{}
	}
	id := report.MakeHostNodeID(report.ExtractHostID(n))
	result := NewDerivedNode(id, n).WithTopology(report.Host)
	result.Latest = result.Latest.Set(report.HostNodeID, timestamp, hostNodeID)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return report.Nodes{id: result}
}

// MapEndpoint2Host takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
func MapEndpoint2Host(n report.Node, local report.Networks) report.Nodes {
	// Nodes without a hostid are treated as pseudo nodes
	hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID)
	if !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	id := report.MakeHostNodeID(report.ExtractHostID(n))
	result := NewDerivedNode(id, n).WithTopology(report.Host)
	result.Latest = result.Latest.Set(report.HostNodeID, timestamp, hostNodeID)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return report.Nodes{id: result}
}
