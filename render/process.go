package render

import (
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	InboundMajor  = "The Internet"
	OutboundMajor = "The Internet"
	InboundMinor  = "Inbound connections"
	OutboundMinor = "Outbound connections"

	// Topology for pseudo-nodes and IPs so we can differentiate them at the end
	Pseudo = "pseudo"
)

func renderProcesses(rpt report.Report) bool {
	return len(rpt.Process.Nodes) >= 1
}

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = SelectEndpoint

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = Memoise(endpoints2Processes{})

// ColorConnectedProcessRenderer colors connected nodes from
// ProcessRenderer. Since the process topology views only show
// connected processes, we need this info to determine whether
// processes appearing in a details panel are linkable.
var ColorConnectedProcessRenderer = Memoise(ColorConnected(ProcessRenderer))

// processWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type processWithContainerNameRenderer struct {
	Renderer
}

func (r processWithContainerNameRenderer) Render(rpt report.Report) Nodes {
	processes := r.Renderer.Render(rpt)
	containers := SelectContainer.Render(rpt)

	outputs := report.Nodes{}
	for id, p := range processes.Nodes {
		outputs[id] = p
		containerID, timestamp, ok := p.Latest.LookupEntry(docker.ContainerID)
		if !ok {
			continue
		}
		container, ok := containers.Nodes[report.MakeContainerNodeID(containerID)]
		if !ok {
			continue
		}
		p.Latest = p.Latest.Set(docker.ContainerID, timestamp, containerID)
		if containerName, timestamp, ok := container.Latest.LookupEntry(docker.ContainerName); ok {
			p.Latest = p.Latest.Set(docker.ContainerName, timestamp, containerName)
		}
		outputs[id] = p
	}
	return Nodes{Nodes: outputs, Filtered: processes.Filtered}
}

// ProcessWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
//
// not memoised
var ProcessWithContainerNameRenderer = processWithContainerNameRenderer{ProcessRenderer}

// ProcessNameRenderer is a Renderer which produces a renderable process
// name graph by munging the progess graph.
//
// not memoised
var ProcessNameRenderer = CustomRenderer{Renderer: ProcessRenderer, RenderFunc: processes2Names}

// endpoints2Processes joins the endpoint topology to the process
// topology, matching on hostID and pid.
type endpoints2Processes struct {
}

func (e endpoints2Processes) Render(rpt report.Report) Nodes {
	if len(rpt.Process.Nodes) == 0 {
		return Nodes{}
	}
	local := LocalNetworks(rpt)
	processes := SelectProcess.Render(rpt)
	endpoints := SelectEndpoint.Render(rpt)
	ret := newJoinResults()

	for _, n := range endpoints.Nodes {
		// Nodes without a hostid are treated as pseudo nodes
		if hostNodeID, ok := n.Latest.Lookup(report.HostNodeID); !ok {
			if id, ok := pseudoNodeID(n, local); ok {
				ret.addChild(n, id, newPseudoNode)
			}
		} else {
			pid, timestamp, ok := n.Latest.LookupEntry(process.PID)
			if !ok {
				continue
			}

			if len(n.Adjacency) > 1 {
				// We cannot be sure that the pid is associated with all the
				// connections. It is better to drop such an endpoint than
				// risk rendering bogus connections.
				continue
			}

			hostID, _, _ := report.ParseNodeID(hostNodeID)
			id := report.MakeProcessNodeID(hostID, pid)
			ret.addChild(n, id, func(id string) report.Node {
				if processNode, found := processes.Nodes[id]; found {
					return processNode
				}
				// we have a pid, but no matching process node; create a new one rather than dropping the data
				return report.MakeNode(id).WithTopology(report.Process).
					WithLatest(process.PID, timestamp, pid)
			})
		}
	}
	ret.copyUnmatched(processes)
	ret.fixupAdjacencies(processes)
	ret.fixupAdjacencies(endpoints)
	return ret.result()
}

// processes2Names maps process Nodes to Nodes for each process name.
func processes2Names(processes Nodes) Nodes {
	ret := newJoinResults()

	for _, n := range processes.Nodes {
		if n.Topology == Pseudo {
			ret.passThrough(n)
		} else {
			name, timestamp, ok := n.Latest.LookupEntry(process.Name)
			if ok {
				ret.addChildAndChildren(n, name, func(id string) report.Node {
					return report.MakeNode(id).WithTopology(MakeGroupNodeTopology(n.Topology, process.Name)).
						WithLatest(process.Name, timestamp, name)
				})
			}
		}
	}
	ret.fixupAdjacencies(processes)
	return ret.result()
}
