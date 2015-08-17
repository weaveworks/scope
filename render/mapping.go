package render

import (
	"fmt"
	"net"
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
)

// LeafMapFunc is anything which can take an arbitrary NodeMetadata, which is
// always one-to-one with nodes in a topology, and return a specific
// representation of the referenced node, in the form of a node ID and a
// human-readable major and minor labels.
//
// A single NodeMetadata can yield arbitrary many representations, including
// representations that reduce the cardinality of the set of nodes.
//
// If the final output parameter is false, the node shall be omitted from the
// rendered topology.
type LeafMapFunc func(report.NodeMetadata) (RenderableNode, bool)

// PseudoFunc creates RenderableNode representing pseudo nodes given the nodeID.
// dstNodeID is the node id of one of the nodes this node is attached to.
// nodeID and dstNodeID are node IDs prior to mapping.  isClient indicated the direction
// of the edge to dstNodeID - true indicates nodeID is the client, false indicates
// nodeID is the server.
type PseudoFunc func(nodeID string, dstNodeID string, isClient bool, local report.Networks) (RenderableNode, bool)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return another RenderableNode.
//
// As with LeafMapFunc, if the final output parameter is false, the node
// shall be omitted from the rendered topology.
type MapFunc func(RenderableNode) (RenderableNode, bool)

// MapEndpointIdentity maps an endpoint topology node to an endpoint
// renderable node. As it is only ever run on endpoint topology nodes, we
// expect that certain keys are present.
func MapEndpointIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNode{}, false
	}

	port, ok := m.Metadata[endpoint.Port]
	if !ok {
		return RenderableNode{}, false
	}

	var (
		id    = MakeEndpointID(report.ExtractHostID(m), addr, port)
		major = fmt.Sprintf("%s:%s", addr, port)
		minor = report.ExtractHostID(m)
		rank  = major
	)

	if pid, ok := m.Metadata[process.PID]; ok {
		minor = fmt.Sprintf("%s (%s)", minor, pid)
	}

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapProcessIdentity maps a process topology node to a process renderable
// node. As it is only ever run on process topology nodes, we expect that
// certain keys are present.
func MapProcessIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	pid, ok := m.Metadata[process.PID]
	if !ok {
		return RenderableNode{}, false
	}

	var (
		id    = MakeProcessID(report.ExtractHostID(m), pid)
		major = m.Metadata["comm"]
		minor = fmt.Sprintf("%s (%s)", report.ExtractHostID(m), pid)
		rank  = m.Metadata["comm"]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapContainerIdentity maps a container topology node to a container
// renderable node. As it is only ever run on container topology nodes, we
// expect that certain keys are present.
func MapContainerIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	id, ok := m.Metadata[docker.ContainerID]
	if !ok {
		return RenderableNode{}, false
	}

	var (
		major = m.Metadata[docker.ContainerName]
		minor = report.ExtractHostID(m)
		rank  = m.Metadata[docker.ImageID]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapContainerImageIdentity maps a container image topology node to container
// image renderable node. As it is only ever run on container image topology
// nodes, we expect that certain keys are present.
func MapContainerImageIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	id, ok := m.Metadata[docker.ImageID]
	if !ok {
		return RenderableNode{}, false
	}

	var (
		major = m.Metadata[docker.ImageName]
		rank  = m.Metadata[docker.ImageID]
	)

	return NewRenderableNode(id, major, "", rank, m), true
}

// MapAddressIdentity maps an address topology node to an address renderable
// node. As it is only ever run on address topology nodes, we expect that
// certain keys are present.
func MapAddressIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNode{}, false
	}

	var (
		id    = MakeAddressID(report.ExtractHostID(m), addr)
		major = addr
		minor = report.ExtractHostID(m)
		rank  = major
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapHostIdentity maps a host topology node to a host renderable node. As it
// is only ever run on host topology nodes, we expect that certain keys are
// present.
func MapHostIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id                 = MakeHostID(report.ExtractHostID(m))
		hostname           = m.Metadata[host.HostName]
		parts              = strings.SplitN(hostname, ".", 2)
		major, minor, rank = "", "", ""
	)

	if len(parts) == 2 {
		major, minor, rank = parts[0], parts[1], parts[1]
	} else {
		major = hostname
	}

	return NewRenderableNode(id, major, minor, rank, m), true
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
func MapEndpoint2Process(n RenderableNode) (RenderableNode, bool) {
	if n.Pseudo {
		return n, true
	}

	pid, ok := n.NodeMetadata.Metadata[process.PID]
	if !ok {
		return RenderableNode{}, false
	}

	id := MakeProcessID(report.ExtractHostID(n.NodeMetadata), pid)
	return newDerivedNode(id, n), true
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
func MapProcess2Container(n RenderableNode) (RenderableNode, bool) {
	// Propogate the internet pseudo node
	if n.ID == TheInternetID {
		return n, true
	}

	// Don't propogate non-internet pseudo nodes
	if n.Pseudo {
		return n, false
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
		return node, true
	}

	return newDerivedNode(id, n), true
}

// MapProcess2Name maps process RenderableNodes to RenderableNodes
// for each process name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapProcess2Name(n RenderableNode) (RenderableNode, bool) {
	if n.Pseudo {
		return n, true
	}

	name, ok := n.NodeMetadata.Metadata["comm"]
	if !ok {
		return RenderableNode{}, false
	}

	node := newDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	return node, true
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
func MapContainer2ContainerImage(n RenderableNode) (RenderableNode, bool) {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return n, true
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.NodeMetadata.Metadata[docker.ImageID]
	if !ok {
		return n, false
	}

	return newDerivedNode(id, n), true
}

// MapContainerImage2Name maps container images RenderableNodes to
// RenderableNodes for each container image name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapContainerImage2Name(n RenderableNode) (RenderableNode, bool) {
	if n.Pseudo {
		return n, true
	}

	name, ok := n.NodeMetadata.Metadata[docker.ImageName]
	if !ok {
		return RenderableNode{}, false
	}

	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		name = parts[0]
	}

	node := newDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	return node, true
}

// MapAddress2Host maps address RenderableNodes to host RenderableNodes.
//
// Otherthan pseudo nodes, we can assume all nodes have a HostID
func MapAddress2Host(n RenderableNode) (RenderableNode, bool) {
	if n.Pseudo {
		return n, true
	}

	id := MakeHostID(report.ExtractHostID(n.NodeMetadata))
	return newDerivedNode(id, n), true
}

// GenericPseudoNode makes a PseudoFunc given an addresser.  The returned
// PseudoFunc will produce Internet pseudo nodes for addresses not in
// the report's local networks.  Otherwise, the returned function will
// produce a single pseudo node per (dst address, src address, src port).
func GenericPseudoNode(addresser func(id string) net.IP) PseudoFunc {
	return func(nodeID, dstNodeId string, isClient bool, local report.Networks) (RenderableNode, bool) {
		// Use the addresser to extract the IP of the missing node
		nodeAddr := addresser(nodeID)
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		if !local.Contains(nodeAddr) {
			return newPseudoNode(TheInternetID, TheInternetMajor, ""), true
		}

		if isClient {
			// If the client node is missing, generate a single pseudo node for every (client ip, server ip, server por)
			serverIP, serverPort := trySplitAddr(dstNodeId)
			outputID := MakePseudoNodeID(nodeAddr.String(), serverIP, serverPort)
			major := nodeAddr.String()
			return newPseudoNode(outputID, major, ""), true
		}

		// Othereise (the server node is missing), generate a pseudo node for every (server ip, server port)
		serverIP, serverPort := trySplitAddr(nodeID)
		outputID := MakePseudoNodeID(serverIP, serverPort)
		if serverPort != "" {
			return newPseudoNode(outputID, serverIP+":"+serverPort, ""), true
		}
		return newPseudoNode(outputID, serverIP, ""), true
	}
}

// PanicPseudoNode just panics; it is for Topologies without edges
func PanicPseudoNode(src, dst string, isClient bool, local report.Networks) (RenderableNode, bool) {
	panic(src)
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
