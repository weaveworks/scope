package render

import (
	"github.com/weaveworks/scope/report"
)

// HostRenderer is a Renderer which produces a renderable host
// graph from the host topology.
//
// not memoised
var HostRenderer = MakeReduce(
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ProcessRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ContainerRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: ContainerImageRenderer},
	CustomRenderer{RenderFunc: nodes2Hosts, Renderer: PodRenderer},
	endpoints2Hosts{},
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
	ret := newJoinResults(nil)

	for _, n := range nodes.Nodes {
		if n.Topology == Pseudo {
			continue // Don't propagate pseudo nodes - we do this in endpoints2Hosts
		}
		isImage := n.Topology == report.ContainerImage
		hostIDs, _ := n.Parents.Lookup(report.Host)
		for _, id := range hostIDs {
			if isImage {
				// We need to treat image nodes specially because they
				// aggregate adjacencies of containers across multiple
				// hosts, and hence mapping these adjacencies to host
				// adjacencies would produce edges that aren't present
				// in reality.
				ret.addUnmappedChild(n, id, func(id string) report.Node {
					return report.MakeNode(id).WithTopology(report.Host)
				})
			} else {
				ret.addChild(n, id, func(id string) report.Node {
					return report.MakeNode(id).WithTopology(report.Host)
				})
			}
		}
	}
	return ret.result(nodes)
}

// endpoints2Hosts takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
type endpoints2Hosts struct {
}

func (e endpoints2Hosts) Render(rpt report.Report) Nodes {
	local := LocalNetworks(rpt)
	hosts := SelectHost.Render(rpt)
	endpoints := SelectEndpoint.Render(rpt)
	ret := newJoinResults(hosts.Nodes)

	for _, n := range endpoints.Nodes {
		// Nodes without a hostid are treated as pseudo nodes
		if hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID); !ok {
			if id, ok := pseudoNodeID(n, local); ok {
				ret.addChild(n, id, newPseudoNode)
			}
		} else {
			id := report.MakeHostNodeID(report.ExtractHostID(n))
			ret.addChild(n, id, func(id string) report.Node {
				// we have a hostNodeID, but no matching host node;
				// create a new one rather than dropping the data
				return report.MakeNode(id).WithTopology(report.Host).
					WithLatest(report.HostNodeID, timestamp, hostNodeID)
			})
		}
	}
	return ret.result(endpoints)
}
