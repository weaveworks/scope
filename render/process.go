package render

import (
	"context"

	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	InboundMajor  = "The Internet"
	OutboundMajor = "The Internet"
	InboundMinor  = "Inbound connections"
	OutboundMinor = "Outbound connections"
)

func renderProcesses(rpt report.Report) bool {
	return len(rpt.Process.Nodes) >= 1
}

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = Memoise(endpoints2Processes{})

// ConnectedProcessRenderer is a Renderer which colors
// connected nodes, so we can apply a filter to show/hide unconnected
// nodes depending on user choice.
//
// not memoised
var ConnectedProcessRenderer = ColorConnected(ProcessRenderer)

// ProcessNameRenderer is a Renderer which produces a renderable
// process name graph by munging the progess graph.
//
// It also colors connected nodes, so we can apply a filter to
// show/hide unconnected nodes depending on user choice.
//
// not memoised
var ProcessNameRenderer = ColorConnected(CustomRenderer{RenderFunc: processes2Names, Renderer: ProcessRenderer})

// endpoints2Processes joins the endpoint topology to the process
// topology, matching on hostID and pid.
type endpoints2Processes struct {
}

func (e endpoints2Processes) Render(ctx context.Context, rpt report.Report) Nodes {
	if len(rpt.Process.Nodes) == 0 {
		return Nodes{}
	}
	endpoints := SelectEndpoint.Render(ctx, rpt).Nodes
	return MapEndpoints(
		func(endpoint report.Node) string {
			pid, ok := endpoint.Latest.Lookup(report.PID)
			if !ok {
				return ""
			}
			if hasMoreThanOneConnection(endpoint, endpoints) {
				return ""
			}
			hostID := report.ExtractHostID(endpoint)
			if hostID == "" {
				return ""
			}
			return report.MakeProcessNodeID(hostID, pid)
		}, report.Process).Render(ctx, rpt)
}

// When there is more than one connection originating from a source
// endpoint, we cannot be sure that its pid is associated with all of
// them, since the source endpoint may have been re-used by a
// different process. See #2665. It is better to drop such an endpoint
// than risk rendering bogus connections.  Aliased connections - when
// all the remote endpoints represent the same logical endpoint, due
// to NATing - are fine though.
func hasMoreThanOneConnection(n report.Node, endpoints report.Nodes) bool {
	if len(n.Adjacency) < 2 {
		return false
	}
	firstRealEndpointID := ""
	for _, endpointID := range n.Adjacency {
		if ep, ok := endpoints[endpointID]; ok {
			if copyID, ok := ep.Latest.Lookup(report.CopyOf); ok {
				endpointID = copyID
			}
		}
		if firstRealEndpointID == "" {
			firstRealEndpointID = endpointID
		} else if firstRealEndpointID != endpointID {
			return true
		}
	}
	return false
}

var processNameTopology = MakeGroupNodeTopology(report.Process, report.Name)

// processes2Names maps process Nodes to Nodes for each process name.
func processes2Names(processes Nodes) Nodes {
	ret := newJoinResults(nil)

	for _, n := range processes.Nodes {
		if n.Topology == Pseudo {
			ret.passThrough(n)
		} else if name, ok := n.Latest.Lookup(report.Name); ok {
			ret.addChildAndChildren(n, name, processNameTopology)
		}
	}
	return ret.result(processes.Nodes)
}
