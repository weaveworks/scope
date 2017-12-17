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
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ProcessRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ContainerRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ContainerImageRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: PodRenderer},
	SelectHost,
)

// nodes2Hosts maps any Nodes to host Nodes.
//
// If this function is given a node without a hostname
// (including other pseudo nodes), it will drop the node.
//
// Otherwise, this function will produce nodes with the correct ID
// format for a host, but without any Major or Minor labels.  It does
// not have enough info to do that, and the resulting graph must be
// merged with a host graph to get that info.
func nodes2Hosts(nodes Nodes) Nodes {
	ret := newJoinResults()

	for _, n := range nodes.Nodes {
		if n.Topology == Pseudo {
			continue // Don't propagate pseudo nodes - we do this in endpoints2Hosts
		}
		hostIDs, _ := n.Parents.Lookup(report.Host)
		for _, id := range hostIDs {
			ret.addChild(n, id, func(id string) report.Node {
				return report.MakeNode(id).WithTopology(report.Host)
			})
		}
	}
	ret.fixupAdjacencies(nodes)
	return ret.result()
}

// endpoints2Hosts takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
type endpoints2Hosts struct {
}

func (e endpoints2Hosts) Render(rpt report.Report) Nodes {
	local := LocalNetworks(rpt)
	endpoints := SelectEndpoint.Render(rpt)
	ret := newJoinResults()

	for _, n := range endpoints.Nodes {
		// Nodes without a hostid are treated as pseudo nodes
		if hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID); !ok {
			if id, ok := pseudoNodeID(n, local); ok {
				ret.addChild(n, id, newPseudoNode)
			}
		} else {
			id := report.MakeHostNodeID(report.ExtractHostID(n))
			ret.addChild(n, id, func(id string) report.Node {
				return report.MakeNode(id).WithTopology(report.Host).
					WithLatest(report.HostNodeID, timestamp, hostNodeID)
			})
		}
	}
	ret.fixupAdjacencies(endpoints)
	return ret.result()
}
