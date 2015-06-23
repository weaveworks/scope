package render

import (
	"fmt"
	"net"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
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

// PseudoFunc creates RenderableNode representing pseudo nodes given the dstNodeID.
// The srcNode renderable node is essentially from MapFunc, representing one of
// the rendered nodes this pseudo node refers to. srcNodeID and dstNodeID are
// node IDs prior to mapping.
type PseudoFunc func(srcNodeID string, srcNode RenderableNode, dstNodeID string, local report.Networks) (RenderableNode, bool)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return another RenderableNode.
//
// As with LeafMapFunc, if the final output parameter is false, the node
// shall be omitted from the rendered topology.
type MapFunc func(RenderableNode) (RenderableNode, bool)

// MapEndpointIdentity maps a endpoint topology node to endpoint RenderableNode
// node. As it is only ever run on endpoint topology nodes, we can safely
// assume the presence of certain keys.
func MapEndpointIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id      = MakeEndpointID(report.ExtractHostID(m), m["addr"], m["port"])
		major   = fmt.Sprintf("%s:%s", m["addr"], m["port"])
		pid, ok = m["pid"]
		minor   = report.ExtractHostID(m)
		rank    = major
	)

	if ok {
		minor = fmt.Sprintf("%s (%s)", report.ExtractHostID(m), pid)
	}

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapProcessIdentity maps a process topology node to process RenderableNode node.
// As it is only ever run on process topology nodes, we can safely assume the
// presence of certain keys.
func MapProcessIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = MakeProcessID(report.ExtractHostID(m), m["pid"])
		major = m["comm"]
		minor = fmt.Sprintf("%s (%s)", report.ExtractHostID(m), m["pid"])
		rank  = m["pid"]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapContainerIdentity maps a container topology node to a container
// RenderableNode node. As it is only ever run on container topology
// nodes, we can safely assume the presences of certain keys.
func MapContainerIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = m[docker.ContainerID]
		major = m[docker.ContainerName]
		minor = report.ExtractHostID(m)
		rank  = m[docker.ImageID]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapContainerImageIdentity maps a container image topology node to container
// image RenderableNode node. As it is only ever run on container image
// topology nodes, we can safely assume the presences of certain keys.
func MapContainerImageIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = m[docker.ImageID]
		major = m[docker.ImageName]
		rank  = m[docker.ImageID]
	)

	return NewRenderableNode(id, major, "", rank, m), true
}

// MapAddressIdentity maps a address topology node to address RenderableNode
// node. As it is only ever run on address topology nodes, we can safely
// assume the presence of certain keys.
func MapAddressIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = MakeAddressID(report.ExtractHostID(m), m["addr"])
		major = m["addr"]
		minor = report.ExtractHostID(m)
		rank  = major
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapHostIdentity maps a host topology node to host RenderableNode
// node. As it is only ever run on host topology nodes, we can safely
// assume the presence of certain keys.
func MapHostIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id                 = MakeHostID(report.ExtractHostID(m))
		hostname           = m[host.HostName]
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

	pid, ok := n.NodeMetadata["pid"]
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
	id, ok := n.NodeMetadata[docker.ContainerID]
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

	name, ok := n.NodeMetadata["comm"]
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
	id, ok := n.NodeMetadata[docker.ImageID]
	if !ok {
		return n, false
	}

	return newDerivedNode(id, n), true
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
	return func(src string, srcMapped RenderableNode, dst string, local report.Networks) (RenderableNode, bool) {
		// Use the addresser to extract the destination IP
		dstNodeAddr := addresser(dst)

		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		if !local.Contains(dstNodeAddr) {
			return newPseudoNode(TheInternetID, TheInternetMajor, ""), true
		}

		// Otherwise, the rule for non-internet psuedo nodes; emit 1 new node for each
		// dstNodeAddr, srcNodeAddr, srcNodePort.
		srcNodeAddr, srcNodePort := trySplitAddr(src)

		outputID := MakePseudoNodeID(dstNodeAddr.String(), srcNodeAddr, srcNodePort)
		major := dstNodeAddr.String()
		return newPseudoNode(outputID, major, ""), true
	}
}

// PanicPseudoNode just panics; it is for Topologies without edges
func PanicPseudoNode(src string, srcMapped RenderableNode, dst string, local report.Networks) (RenderableNode, bool) {
	panic(dst)
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
