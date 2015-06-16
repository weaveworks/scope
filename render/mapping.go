package render

import (
	"fmt"
	"strings"

	"github.com/weaveworks/scope/report"
)

const (
	uncontainedID    = "uncontained"
	uncontainedMajor = "Uncontained"

	humanTheInternet = "the Internet"
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
type PseudoFunc func(srcNodeID string, srcNode RenderableNode, dstNodeID string) (RenderableNode, bool)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return another RenderableNode.
type MapFunc func(RenderableNode) (RenderableNode, bool)

// MapEndpointIdentity maps a endpoint topology node to endpoint RenderableNode node.
// As it is only ever run on endpoint topology nodes, we can safely assume the
// presences of certain keys.
func MapEndpointIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = fmt.Sprintf("endpoint:%s:%s:%s", m[report.HostNodeID], m["addr"], m["port"])
		major = fmt.Sprintf("%s:%s", m["addr"], m["port"])
		minor = fmt.Sprintf("%s (%s)", m[report.HostNodeID], m["pid"])
		rank  = m["pid"]
	)
	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapProcessIdentity maps a process topology node to process RenderableNode node.
// As it is only ever run on process topology nodes, we can safely assume the
// presences of certain keys.
func MapProcessIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = fmt.Sprintf("pid:%s:%s", m[report.HostNodeID], m["pid"])
		major = m["comm"]
		minor = fmt.Sprintf("%s (%s)", m[report.HostNodeID], m["pid"])
		rank  = m["pid"]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapContainerIdentity maps a container topology node to container RenderableNode node.
// As it is only ever run on container topology nodes, we can safely assume the
// presences of certain keys.
func MapContainerIdentity(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		id    = m["docker_container_id"]
		major = m["docker_container_name"]
		minor = m[report.HostNodeID]
		rank  = m["docker_image_id"]
	)

	return NewRenderableNode(id, major, minor, rank, m), true
}

// MapEndpoint2Process maps endpoint RenderableNodes to process RenderableNodes.
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

	id := fmt.Sprintf("pid:%s:%s", n.NodeMetadata[report.HostNodeID], n.NodeMetadata["pid"])
	return newDerivedNode(id, n), true
}

// MapProcess2Container maps process RenderableNodes to container RenderableNodes.
//
// If this function is given a node without a docker_container_id (including other
// psueod nodes), it will produce an "Uncontained" pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapProcess2Container(n RenderableNode) (RenderableNode, bool) {
	id, ok := n.NodeMetadata["docker_container_id"]
	if !ok {
		return newPseudoNode(uncontainedID, uncontainedMajor, ""), true
	}

	return newDerivedNode(id, n), true
}

// ProcessPID takes a node NodeMetadata from topology, and returns a
// representation with the ID based on the process PID and the labels based on
// the process name.
func ProcessPID(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		identifier = fmt.Sprintf("%s:%s:%s", "pid", m["domain"], m["pid"])
		minor      = fmt.Sprintf("%s (%s)", m["domain"], m["pid"])
		show       = m["pid"] != "" && m["name"] != ""
	)

	return NewRenderableNode(identifier, m["name"], minor, m["pid"], m), show
}

// ProcessName takes a node NodeMetadata from a topology, and returns a
// representation with the ID based on the process name (grouping all
// processes with the same name together).
func ProcessName(m report.NodeMetadata) (RenderableNode, bool) {
	show := m["pid"] != "" && m["name"] != ""
	return NewRenderableNode(m["name"], m["name"], "", m["name"], m), show
}

// ProcessContainerImage maps topology nodes to the container images they run
// on. If no container metadata is found, nodes are grouped into the
// Uncontained node.
func ProcessContainerImage(m report.NodeMetadata) (RenderableNode, bool) {
	var id, major, minor, rank string
	if m["docker_image_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_image_id"], m["docker_image_name"], "", m["docker_image_id"]
	}

	return NewRenderableNode(id, major, minor, rank, m), true
}

// NetworkHostname takes a node NodeMetadata and returns a representation
// based on the hostname. Major label is the hostname, the minor label is the
// domain, if any.
func NetworkHostname(m report.NodeMetadata) (RenderableNode, bool) {
	var (
		name   = m["name"]
		domain = ""
		parts  = strings.SplitN(name, ".", 2)
	)

	if len(parts) == 2 {
		domain = parts[1]
	}

	return NewRenderableNode(fmt.Sprintf("host:%s", name), parts[0], domain, parts[0], m), name != ""
}

// GenericPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericPseudoNode(src string, srcMapped RenderableNode, dst string) (RenderableNode, bool) {
	var maj, min, outputID string

	if dst == report.TheInternet {
		outputID = dst
		maj, min = humanTheInternet, ""
	} else {
		// Rule for non-internet psuedo nodes; emit 1 new node for each
		// dstNodeAddr, srcNodeAddr, srcNodePort.
		srcNodeAddr, srcNodePort := trySplitAddr(src)
		dstNodeAddr, _ := trySplitAddr(dst)

		outputID = report.MakePseudoNodeID(dstNodeAddr, srcNodeAddr, srcNodePort)
		maj, min = dstNodeAddr, ""
	}

	return newPseudoNode(outputID, maj, min), true
}

// GenericGroupedPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericGroupedPseudoNode(src string, srcMapped RenderableNode, dst string) (RenderableNode, bool) {
	var maj, min, outputID string

	if dst == report.TheInternet {
		outputID = dst
		maj, min = humanTheInternet, ""
	} else {
		// When grouping, emit one pseudo node per (srcNodeAddress, dstNodeAddr)
		dstNodeAddr, _ := trySplitAddr(dst)

		outputID = report.MakePseudoNodeID(dstNodeAddr, srcMapped.ID)
		maj, min = dstNodeAddr, ""
	}

	return newPseudoNode(outputID, maj, min), true
}

// InternetOnlyPseudoNode never creates a pseudo node, unless it's the Internet.
func InternetOnlyPseudoNode(_ string, _ RenderableNode, dst string) (RenderableNode, bool) {
	if dst == report.TheInternet {
		return newPseudoNode(report.TheInternet, humanTheInternet, ""), true
	}
	return RenderableNode{}, false
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
