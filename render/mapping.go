package render

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UncontainedID    = "uncontained"
	UncontainedMajor = "Uncontained"

	TheInternetID    = "theinternet"
	TheInternetMajor = "The Internet"

	containersKey = "containers"
	processesKey  = "processes"
)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return a set of other RenderableNodes.
//
// As with LeafMapFunc, if the output is empty, the node
// shall be omitted from the rendered topology.
type MapFunc func(RenderableNode, report.Networks) RenderableNodes

// MapEndpointIdentity maps an endpoint topology node to a single endpoint
// renderable node. As it is only ever run on endpoint topology nodes, we
// expect that certain keys are present.
func MapEndpointIdentity(m RenderableNode, local report.Networks) RenderableNodes {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}

	port, ok := m.Metadata[endpoint.Port]
	if !ok {
		return RenderableNodes{}
	}

	// Nodes without a hostid are treated as psuedo nodes
	_, ok = m.Metadata[report.HostNodeID]
	if !ok {
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		if !local.Contains(net.ParseIP(addr)) {
			return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
		}

		// We are a 'client' pseudo node if the port is in the ephemeral port range.
		// Linux uses 32768 to 61000.
		if p, err := strconv.Atoi(port); err == nil && len(m.Adjacency) > 0 && p >= 32768 && p < 61000 {
			// We only exist if there is something in our adjacency
			// Generate a single pseudo node for every (client ip, server ip, server port)
			dstNodeID := m.Adjacency[0]
			serverIP, serverPort := trySplitAddr(dstNodeID)
			outputID := MakePseudoNodeID(addr, serverIP, serverPort)
			return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
		}

		// Otherwise (the server node is missing), generate a pseudo node for every (server ip, server port)
		outputID := MakePseudoNodeID(addr, port)
		if port != "" {
			return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr+":"+port, m)}
		}
		return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
	}

	var (
		id    = MakeEndpointID(report.ExtractHostID(m.NodeMetadata), addr, port)
		major = fmt.Sprintf("%s:%s", addr, port)
		minor = report.ExtractHostID(m.NodeMetadata)
		rank  = major
	)

	if pid, ok := m.Metadata[process.PID]; ok {
		minor = fmt.Sprintf("%s (%s)", minor, pid)
	}

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapProcessIdentity maps a process topology node to a process renderable
// node. As it is only ever run on process topology nodes, we expect that
// certain keys are present.
func MapProcessIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	pid, ok := m.Metadata[process.PID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		id    = MakeProcessID(report.ExtractHostID(m.NodeMetadata), pid)
		major = m.Metadata["comm"]
		minor = fmt.Sprintf("%s (%s)", report.ExtractHostID(m.NodeMetadata), pid)
		rank  = m.Metadata["comm"]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapContainerIdentity maps a container topology node to a container
// renderable node. As it is only ever run on container topology nodes, we
// expect that certain keys are present.
func MapContainerIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[docker.ContainerID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major = m.Metadata[docker.ContainerName]
		minor = report.ExtractHostID(m.NodeMetadata)
		rank  = m.Metadata[docker.ImageID]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapContainerImageIdentity maps a container image topology node to container
// image renderable node. As it is only ever run on container image topology
// nodes, we expect that certain keys are present.
func MapContainerImageIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[docker.ImageID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major = m.Metadata[docker.ImageName]
		rank  = m.Metadata[docker.ImageID]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, "", rank, m)}
}

// MapAddressIdentity maps an address topology node to an address renderable
// node. As it is only ever run on address topology nodes, we expect that
// certain keys are present.
func MapAddressIdentity(m RenderableNode, local report.Networks) RenderableNodes {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}

	// Nodes without a hostid are treated as psuedo nodes
	_, ok = m.Metadata[report.HostNodeID]
	if !ok {
		// If the addr is not in a network local to this report, we emit an
		// internet node
		if !local.Contains(net.ParseIP(addr)) {
			return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
		}

		// Otherwise generate a pseudo node for every
		outputID := MakePseudoNodeID(addr, "")
		if len(m.Adjacency) > 0 {
			_, dstAddr, _ := report.ParseAddressNodeID(m.Adjacency[0])
			outputID = MakePseudoNodeID(addr, dstAddr)
		}
		return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
	}

	var (
		id    = MakeAddressID(report.ExtractHostID(m.NodeMetadata), addr)
		major = addr
		minor = report.ExtractHostID(m.NodeMetadata)
		rank  = major
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapHostIdentity maps a host topology node to a host renderable node. As it
// is only ever run on host topology nodes, we expect that certain keys are
// present.
func MapHostIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	var (
		id                 = MakeHostID(report.ExtractHostID(m.NodeMetadata))
		hostname           = m.Metadata[host.HostName]
		parts              = strings.SplitN(hostname, ".", 2)
		major, minor, rank = "", "", ""
	)

	if len(parts) == 2 {
		major, minor, rank = parts[0], parts[1], parts[1]
	} else {
		major = hostname
	}

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapEndpoint2IP maps endpoint nodes to their IP address, for joining
// with container nodes.  We drop endpoint nodes with pids, as they
// will be joined to containers through the process topology, and we
// don't want to double count edges.
func MapEndpoint2IP(m RenderableNode, local report.Networks) RenderableNodes {
	_, ok := m.Metadata[process.PID]
	if ok {
		return RenderableNodes{}
	}
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}
	if !local.Contains(net.ParseIP(addr)) {
		return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
	}
	return RenderableNodes{addr: NewRenderableNodeWith(addr, "", "", "", m)}
}

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m RenderableNode, _ report.Networks) RenderableNodes {
	result := RenderableNodes{}
	addrs, ok := m.Metadata[docker.ContainerIPs]
	if !ok {
		return result
	}
	for _, addr := range strings.Fields(addrs) {
		n := NewRenderableNodeWith(addr, "", "", "", m)
		n.NodeMetadata.Counters[containersKey] = 1
		result[addr] = n
	}
	return result
}

// MapIP2Container maps IP nodes produced from MapContainer2IP back to
// container nodes.  If there is more than one container with a given
// IP, it is dropped.
func MapIP2Container(n RenderableNode, _ report.Networks) RenderableNodes {
	// If an IP is shared between multiple containers, we can't
	// reliably attribute an connection based on its IP
	if n.NodeMetadata.Counters[containersKey] > 1 {
		return RenderableNodes{}
	}

	// Propogate the internet pseudo node.
	if n.ID == TheInternetID {
		return RenderableNodes{n.ID: n}
	}

	// If this node is not a container, exclude it.
	// This excludes all the nodes we've dragged in from endpoint
	// that we failed to join to a container.
	id, ok := n.NodeMetadata.Metadata[docker.ContainerID]
	if !ok {
		return RenderableNodes{}
	}

	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapEndpoint2Process maps endpoint RenderableNodes to process
// RenderableNodes.
//
// If this function is given a pseudo node, then it will just return it;
// Pseudo nodes will never have pids in them, and therefore will never
// be able to be turned into a Process node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a process, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a process graph to get that info.
func MapEndpoint2Process(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	pid, ok := n.NodeMetadata.Metadata[process.PID]
	if !ok {
		return RenderableNodes{}
	}

	id := MakeProcessID(report.ExtractHostID(n.NodeMetadata), pid)
	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapProcess2Container maps process RenderableNodes to container
// RenderableNodes.
//
// If this function is given a node without a docker_container_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapProcess2Container(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate the internet pseudo node
	if n.ID == TheInternetID {
		return RenderableNodes{n.ID: n}
	}

	// Don't propogate non-internet pseudo nodes
	if n.Pseudo {
		return RenderableNodes{}
	}

	// Otherwise, if the process is not in a container, group it
	// into an per-host "Uncontained" node.  If for whatever reason
	// this node doesn't have a host id in their nodemetadata, it'll
	// all get grouped into a single uncontained node.
	id, ok := n.NodeMetadata.Metadata[docker.ContainerID]
	if !ok {
		hostID := report.ExtractHostID(n.NodeMetadata)
		id = MakePseudoNodeID(UncontainedID, hostID)
		node := newDerivedPseudoNode(id, UncontainedMajor, n)
		node.LabelMinor = hostID
		return RenderableNodes{id: node}
	}

	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapProcess2Name maps process RenderableNodes to RenderableNodes
// for each process name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapProcess2Name(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	name, ok := n.NodeMetadata.Metadata["comm"]
	if !ok {
		return RenderableNodes{}
	}

	node := NewDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	node.NodeMetadata.Counters[processesKey] = 1
	return RenderableNodes{name: node}
}

// MapCountProcessName maps 1:1 process name nodes, counting
// the number of processes grouped together and putting
// that info in the minor label.
func MapCountProcessName(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	processes := n.NodeMetadata.Counters[processesKey]
	if processes == 1 {
		n.LabelMinor = "1 process"
	} else {
		n.LabelMinor = fmt.Sprintf("%d processes", processes)
	}
	return RenderableNodes{n.ID: n}
}

// MapContainer2ContainerImage maps container RenderableNodes to container
// image RenderableNodes.
//
// If this function is given a node without a docker_image_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2ContainerImage(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.NodeMetadata.Metadata[docker.ImageID]
	if !ok {
		return RenderableNodes{}
	}

	// Add container-<id> key to NMD, which will later be counted to produce the minor label
	result := NewDerivedNode(id, n)
	result.NodeMetadata.Counters[containersKey] = 1
	return RenderableNodes{id: result}
}

// MapContainerImage2Name maps container images RenderableNodes to
// RenderableNodes for each container image name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapContainerImage2Name(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	name, ok := n.NodeMetadata.Metadata[docker.ImageName]
	if !ok {
		return RenderableNodes{}
	}

	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		name = parts[0]
	}

	node := NewDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	node.NodeMetadata = n.NodeMetadata.Copy() // Propagate NMD for container counting.
	return RenderableNodes{name: node}
}

// MapCountContainers maps 1:1 container image nodes, counting
// the number of containers grouped together and putting
// that info in the minor label.
func MapCountContainers(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	containers := n.NodeMetadata.Counters[containersKey]
	if containers == 1 {
		n.LabelMinor = "1 container"
	} else {
		n.LabelMinor = fmt.Sprintf("%d containers", containers)
	}
	return RenderableNodes{n.ID: n}
}

// MapAddress2Host maps address RenderableNodes to host RenderableNodes.
//
// Otherthan pseudo nodes, we can assume all nodes have a HostID
func MapAddress2Host(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	id := MakeHostID(report.ExtractHostID(n.NodeMetadata))
	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// trySplitAddr is basically ParseArbitraryNodeID, since its callsites
// (pseudo funcs) just have opaque node IDs and don't know what topology they
// come from. Without changing how pseudo funcs work, we can't make it much
// smarter.
//
// TODO change how pseudofuncs work, and eliminate this helper.
func trySplitAddr(addr string) (string, string) {
	fields := strings.SplitN(addr, report.ScopeDelim, 3)
	if len(fields) == 3 {
		return fields[1], fields[2]
	}
	if len(fields) == 2 {
		return fields[1], ""
	}
	panic(addr)
}
