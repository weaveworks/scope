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

func (r processWithContainerNameRenderer) Render(rpt report.Report, dct Decorator) Nodes {
	processes := r.Renderer.Render(rpt, dct)
	containers := SelectContainer.Render(rpt, dct)

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
var ProcessNameRenderer = ConditionalRenderer(renderProcesses,
	MakeMap(
		MapProcess2Name,
		ProcessRenderer,
	),
)

// endpoints2Processes joins the endpoint topology to the process
// topology, matching on hostID and pid.
type endpoints2Processes struct {
}

func (e endpoints2Processes) Render(rpt report.Report, dct Decorator) Nodes {
	if len(rpt.Process.Nodes) == 0 {
		return Nodes{}
	}
	local := LocalNetworks(rpt)
	processes := SelectProcess.Render(rpt, dct)
	endpoints := SelectEndpoint.Render(rpt, dct)
	ret := newJoinResults()

	for _, n := range endpoints.Nodes {
		// Nodes without a hostid are treated as pseudo nodes
		if hostNodeID, ok := n.Latest.Lookup(report.HostNodeID); !ok {
			if id, ok := pseudoNodeID(n, local); ok {
				ret.addToResults(n, id, newPseudoNode)
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
			ret.addToResults(n, id, func(id string) report.Node {
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

// MapProcess2Name maps process Nodes to Nodes
// for each process name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.
func MapProcess2Name(n report.Node, _ report.Networks) report.Nodes {
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	name, timestamp, ok := n.Latest.LookupEntry(process.Name)
	if !ok {
		return report.Nodes{}
	}

	node := NewDerivedNode(name, n).WithTopology(MakeGroupNodeTopology(n.Topology, process.Name))
	node.Latest = node.Latest.Set(process.Name, timestamp, name)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{name: node}
}
