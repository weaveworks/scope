package render

import (
	"github.com/weaveworks/scope/report"
)

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology.
//
// not memoised
var HostRenderer = MakeReduce(
	endpoints2Hosts{},
	MakeMap(
		MapX2Host,
		ColorConnectedProcessRenderer,
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
// Otherwise, this function will produce nodes with the correct ID
// format for a host, but without any Major or Minor labels.  It does
// not have enough info to do that, and the resulting graph must be
// merged with a host graph to get that info.
func MapX2Host(n report.Node, _ report.Networks) report.Nodes {
	// Don't propagate pseudo nodes - we do this in MapEndpoint2Host
	if n.Topology == Pseudo {
		return report.Nodes{}
	}

	hostIDs, ok := n.Parents.Lookup(report.Host)
	if !ok {
		return report.Nodes{}
	}

	result := report.Nodes{}
	children := report.MakeNodeSet(n)
	for _, id := range hostIDs {
		node := NewDerivedNode(id, n).WithTopology(report.Host)
		node.Counters = node.Counters.Add(n.Topology, 1)
		node.Children = children
		result[id] = node
	}

	return result
}

// endpoints2Hosts takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
type endpoints2Hosts struct {
}

func (e endpoints2Hosts) Render(rpt report.Report, dct Decorator) report.Nodes {
	ns := SelectEndpoint.Render(rpt, dct)
	local := LocalNetworks(rpt)

	var ret = make(report.Nodes)
	var mapped = map[string]string{} // input node ID -> output node ID
	for _, n := range ns {
		var result report.Node
		var exists bool
		// Nodes without a hostid are treated as pseudo nodes
		hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID)
		if !ok {
			id, ok := pseudoNodeID(n, local)
			if !ok {
				continue
			}
			result, exists = ret[id]
			if !exists {
				result = report.MakeNode(id).WithTopology(Pseudo)
			}
		} else {
			id := report.MakeHostNodeID(report.ExtractHostID(n))
			result, exists = ret[id]
			if !exists {
				result = report.MakeNode(id).WithTopology(report.Host)
				result.Latest = result.Latest.Set(report.HostNodeID, timestamp, hostNodeID)
			}
			result.Children = result.Children.Merge(n.Children)
		}
		result.Children = result.Children.Add(n)
		result.Counters = result.Counters.Add(n.Topology, 1)
		ret[result.ID] = result
		mapped[n.ID] = result.ID
	}
	fixupAdjancencies(ns, ret, mapped)
	return ret
}

// Add Node M to the result set ret under id, creating a new result
// node if not already there, and updating the old-id to new-id mapping
// Note we do not update any counters for child topologies here
func addToResults(m report.Node, id string, ret report.Nodes, mapped map[string]string, create func() report.Node) {
	result, exists := ret[id]
	if !exists {
		result = create()
	}
	result.Children = result.Children.Add(m)
	result.Children = result.Children.Merge(m.Children)
	ret[result.ID] = result
	mapped[m.ID] = result.ID
}

// Rewrite Adjacency for new nodes in ret, original nodes in input, and mapping old->new IDs in mapped
func fixupAdjancencies(input, ret report.Nodes, mapped map[string]string) {
	for _, n := range input {
		outID, ok := mapped[n.ID]
		if !ok {
			continue
		}
		out := ret[outID]
		// for each adjacency in the original node, find out what it maps to (if any),
		// and add that to the new node
		for _, a := range n.Adjacency {
			if mappedDest, found := mapped[a]; found {
				out.Adjacency = out.Adjacency.Add(mappedDest)
			}
		}
		ret[outID] = out
	}
}

func (e endpoints2Hosts) Stats(rpt report.Report, _ Decorator) Stats {
	return Stats{} // nothing to report
}
