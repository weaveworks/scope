package render

import (
	"net"
	"sort"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	TheInternetID      = "theinternet"
	IncomingInternetID = "in-" + TheInternetID
	OutgoingInternetID = "out-" + TheInternetID
	InboundMajor       = "The Internet"
	OutboundMajor      = "The Internet"
	InboundMinor       = "Inbound connections"
	OutboundMinor      = "Outbound connections"

	// Topology for pseudo-nodes and IPs so we can differentiate them at the end
	Pseudo = "pseudo"
)

func renderProcesses(rpt report.Report) bool {
	return len(rpt.Process.Nodes) >= 1
}

// EndpointRenderer is a Renderer which produces a renderable endpoint graph.
var EndpointRenderer = FilterNonProcspied(SelectEndpoint)

// ProcessRenderer is a Renderer which produces a renderable process
// graph by merging the endpoint graph and the process topology.
var ProcessRenderer = ConditionalRenderer(renderProcesses,
	ApplyDecorators(ColorConnected(MakeReduce(
		MakeMap(
			MapEndpoint2Process,
			EndpointRenderer,
		),
		SelectProcess,
	))),
)

// processWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
type processWithContainerNameRenderer struct {
	Renderer
}

func (r processWithContainerNameRenderer) Render(rpt report.Report, dct Decorator) report.Nodes {
	processes := r.Renderer.Render(rpt, dct)
	containers := SelectContainer.Render(rpt, dct)

	outputs := report.Nodes{}
	for id, p := range processes {
		outputs[id] = p
		containerID, timestamp, ok := p.Latest.LookupEntry(docker.ContainerID)
		if !ok {
			continue
		}
		container, ok := containers[report.MakeContainerNodeID(containerID)]
		if !ok {
			continue
		}
		p.Latest = p.Latest.Set(docker.ContainerID, timestamp, containerID)
		if containerName, timestamp, ok := container.Latest.LookupEntry(docker.ContainerName); ok {
			p.Latest = p.Latest.Set(docker.ContainerName, timestamp, containerName)
		}
		outputs[id] = p
	}
	return outputs
}

// ProcessWithContainerNameRenderer is a Renderer which produces a process
// graph enriched with container names where appropriate
var ProcessWithContainerNameRenderer = processWithContainerNameRenderer{ProcessRenderer}

// ProcessNameRenderer is a Renderer which produces a renderable process
// name graph by munging the progess graph.
var ProcessNameRenderer = ConditionalRenderer(renderProcesses,
	MakeMap(
		MapProcess2Name,
		ProcessRenderer,
	),
)

// MapEndpoint2Pseudo makes internet of host pesudo nodes from a endpoint node.
func MapEndpoint2Pseudo(n report.Node, local report.Networks) report.Nodes {
	var node report.Node

	addr, ok := n.Latest.Lookup(endpoint.Addr)
	if !ok {
		return report.Nodes{}
	}

	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// external pseudoNode
		node = externalNode(n)
	} else {
		// due to https://github.com/weaveworks/scope/issues/1323 we are dropping
		// all non-internet pseudo nodes for now.
		// node = NewDerivedPseudoNode(MakePseudoNodeID(addr), n)
		return report.Nodes{}
	}
	return report.Nodes{node.ID: node}
}

// MapEndpoint2Process maps endpoint Nodes to process
// Nodes.
//
// If this function is given a pseudo node, then it will just return it;
// Pseudo nodes will never have pids in them, and therefore will never
// be able to be turned into a Process node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a process, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a process graph to get that info.
func MapEndpoint2Process(n report.Node, local report.Networks) report.Nodes {
	// Nodes without a hostid are treated as pseudo nodes
	if _, ok := n.Latest.Lookup(report.HostNodeID); !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	pid, timestamp, ok := n.Latest.LookupEntry(process.PID)
	if !ok {
		return report.Nodes{}
	}

	id := report.MakeProcessNodeID(report.ExtractHostID(n), pid)
	node := NewDerivedNode(id, n).WithTopology(report.Process)
	node.Latest = node.Latest.Set(process.PID, timestamp, pid)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
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

func externalNode(m report.Node) report.Node {
	// First, check if it's a known service and emit a
	// a specific node if it is
	snoopedHostnames, _ := m.Sets.Lookup(endpoint.SnoopedDNSNames)
	reverseHostnames, _ := m.Sets.Lookup(endpoint.ReverseDNSNames)
	// Sort the names to make the lookup more deterministic
	sort.StringSlice(snoopedHostnames).Sort()
	sort.StringSlice(reverseHostnames).Sort()
	// Intentionally prioritize snooped hostnames
	for _, hostname := range append(snoopedHostnames, reverseHostnames...) {
		if isKnownService(hostname) {
			return NewDerivedPseudoNode(ServiceNodeIDPrefix+hostname, m)
		}
	}

	// emit one internet node for incoming, one for outgoing
	if len(m.Adjacency) > 0 {
		return NewDerivedPseudoNode(IncomingInternetID, m)
	}
	return NewDerivedPseudoNode(OutgoingInternetID, m)
}
