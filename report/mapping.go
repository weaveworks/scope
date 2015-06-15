package report

import (
	"fmt"
	"strings"
)

const humanTheInternet = "the Internet"

func newRenderableNode(id, major, minor, rank string) RenderableNode {
	return RenderableNode{
		ID:         id,
		LabelMajor: major,
		LabelMinor: minor,
		Rank:       rank,
		Pseudo:     false,
		Metadata:   AggregateMetadata{},
	}
}

func newPseudoNode(id, major, minor string) RenderableNode {
	return RenderableNode{
		ID:         id,
		LabelMajor: major,
		LabelMinor: minor,
		Rank:       "",
		Pseudo:     true,
		Metadata:   AggregateMetadata{},
	}
}

// MapFunc is anything which can take an arbitrary NodeMetadata, which is
// always one-to-one with nodes in a topology, and return a specific
// representation of the referenced node, in the form of a node ID and a
// human-readable major and minor labels.
//
// A single NodeMetadata can yield arbitrary many representations, including
// representations that reduce the cardinality of the set of nodes.
//
// If the final output parameter is false, the node shall be omitted from the
// rendered topology.
type MapFunc func(NodeMetadata) (RenderableNode, bool)

// PseudoFunc creates RenderableNode representing pseudo nodes given the dstNodeID.
// The srcNode renderable node is essentially from MapFunc, representing one of
// the rendered nodes this pseudo node refers to. srcNodeID and dstNodeID are
// node IDs prior to mapping.
type PseudoFunc func(srcNodeID string, srcNode RenderableNode, dstNodeID string) (RenderableNode, bool)

// TopologySelector selects a single topology from a report.
type TopologySelector func(r Report) Topology

// SelectEndpoint selects the endpoint topology.
func SelectEndpoint(r Report) Topology {
	return r.Endpoint
}

// SelectAddress selects the address topology.
func SelectAddress(r Report) Topology {
	return r.Address
}

// SelectContainer selects the container topology.
func SelectContainer(r Report) Topology {
	return r.Container
}

// ProcessPID takes a node NodeMetadata from topology, and returns a
// representation with the ID based on the process PID and the labels based on
// the process name.
func ProcessPID(m NodeMetadata) (RenderableNode, bool) {
	var (
		identifier = fmt.Sprintf("%s:%s:%s", "pid", m["domain"], m["pid"])
		minor      = fmt.Sprintf("%s (%s)", m["domain"], m["pid"])
		show       = m["pid"] != "" && m["name"] != ""
	)

	return newRenderableNode(identifier, m["name"], minor, m["pid"]), show
}

// ProcessName takes a node NodeMetadata from a topology, and returns a
// representation with the ID based on the process name (grouping all
// processes with the same name together).
func ProcessName(m NodeMetadata) (RenderableNode, bool) {
	show := m["pid"] != "" && m["name"] != ""
	return newRenderableNode(m["name"], m["name"], "", m["name"]), show
}

// MapEndpoint2Container maps endpoint topology nodes to the containers they run
// in. We consider container and image IDs to be globally unique, and so don't
// scope them further by e.g. host. If no container metadata is found, nodes are
// grouped into the Uncontained node.
func MapEndpoint2Container(m NodeMetadata) (RenderableNode, bool) {
	var id, major, minor, rank string
	if m["docker_container_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_container_id"], "", m["domain"], ""
	}

	return newRenderableNode(id, major, minor, rank), true
}

// MapContainerIdentity maps container topology node to container mapped nodes.
func MapContainerIdentity(m NodeMetadata) (RenderableNode, bool) {
	var id, major, minor, rank string
	if m["docker_container_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_container_id"], m["docker_container_name"], m["domain"], m["docker_image_id"]
	}

	return newRenderableNode(id, major, minor, rank), true
}

// ProcessContainerImage maps topology nodes to the container images they run
// on. If no container metadata is found, nodes are grouped into the
// Uncontained node.
func ProcessContainerImage(m NodeMetadata) (RenderableNode, bool) {
	var id, major, minor, rank string
	if m["docker_image_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_image_id"], m["docker_image_name"], "", m["docker_image_id"]
	}

	return newRenderableNode(id, major, minor, rank), true
}

// NetworkHostname takes a node NodeMetadata and returns a representation
// based on the hostname. Major label is the hostname, the minor label is the
// domain, if any.
func NetworkHostname(m NodeMetadata) (RenderableNode, bool) {
	var (
		name   = m["name"]
		domain = ""
		parts  = strings.SplitN(name, ".", 2)
	)

	if len(parts) == 2 {
		domain = parts[1]
	}

	return newRenderableNode(fmt.Sprintf("host:%s", name), parts[0], domain, parts[0]), name != ""
}

// GenericPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericPseudoNode(src string, srcMapped RenderableNode, dst string) (RenderableNode, bool) {
	var maj, min, outputID string

	if dst == TheInternet {
		outputID = dst
		maj, min = humanTheInternet, ""
	} else {
		// Rule for non-internet psuedo nodes; emit 1 new node for each
		// dstNodeAddr, srcNodeAddr, srcNodePort.
		srcNodeAddr, srcNodePort := trySplitAddr(src)
		dstNodeAddr, _ := trySplitAddr(dst)

		outputID = MakePseudoNodeID(dstNodeAddr, srcNodeAddr, srcNodePort)
		maj, min = dstNodeAddr, ""
	}

	return newPseudoNode(outputID, maj, min), true
}

// GenericGroupedPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericGroupedPseudoNode(src string, srcMapped RenderableNode, dst string) (RenderableNode, bool) {
	var maj, min, outputID string

	if dst == TheInternet {
		outputID = dst
		maj, min = humanTheInternet, ""
	} else {
		// When grouping, emit one pseudo node per (srcNodeAddress, dstNodeAddr)
		dstNodeAddr, _ := trySplitAddr(dst)

		outputID = MakePseudoNodeID(dstNodeAddr, srcMapped.ID)
		maj, min = dstNodeAddr, ""
	}

	return newPseudoNode(outputID, maj, min), true
}

// InternetOnlyPseudoNode never creates a pseudo node, unless it's the Internet.
func InternetOnlyPseudoNode(_ string, _ RenderableNode, dst string) (RenderableNode, bool) {
	if dst == TheInternet {
		return newPseudoNode(TheInternet, humanTheInternet, ""), true
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
	fields := strings.SplitN(addr, ScopeDelim, 3)
	if len(fields) == 3 {
		return fields[1], fields[2]
	}
	if len(fields) == 2 {
		return fields[1], ""
	}
	panic(addr)
}
