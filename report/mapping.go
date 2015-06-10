package report

import (
	"fmt"
	"strings"
)

const humanTheInternet = "the Internet"

// MappedNode is returned by the MapFuncs.
type MappedNode struct {
	ID    string
	Major string
	Minor string
	Rank  string
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
type MapFunc func(string, NodeMetadata) (MappedNode, bool)

// PseudoFunc creates MappedNode representing pseudo nodes given the dstNodeID.
// The srcNode renderable node is essentially from MapFunc, representing one of
// the rendered nodes this pseudo node refers to. srcNodeID and dstNodeID are
// node IDs prior to mapping.
type PseudoFunc func(srcNodeID string, srcNode RenderableNode, dstNodeID string) (MappedNode, bool)

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

// ProcessPID takes a node NodeMetadata from topology, and returns a
// representation with the ID based on the process PID and the labels based on
// the process name.
func ProcessPID(_ string, m NodeMetadata) (MappedNode, bool) {
	var (
		identifier = fmt.Sprintf("%s:%s:%s", "pid", m["domain"], m["pid"])
		minor      = fmt.Sprintf("%s (%s)", m["domain"], m["pid"])
		show       = m["pid"] != "" && m["name"] != ""
	)

	return MappedNode{
		ID:    identifier,
		Major: m["name"],
		Minor: minor,
		Rank:  m["pid"],
	}, show
}

// ProcessName takes a node NodeMetadata from a topology, and returns a
// representation with the ID based on the process name (grouping all
// processes with the same name together).
func ProcessName(_ string, m NodeMetadata) (MappedNode, bool) {
	show := m["pid"] != "" && m["name"] != ""
	return MappedNode{
		ID:    m["name"],
		Major: m["name"],
		Minor: "",
		Rank:  m["name"],
	}, show
}

// ProcessContainer maps topology nodes to the containers they run in. We
// consider container and image IDs to be globally unique, and so don't scope
// them further by e.g. host. If no container metadata is found, nodes are
// grouped into the Uncontained node.
func ProcessContainer(_ string, m NodeMetadata) (MappedNode, bool) {
	var id, major, minor, rank string
	if m["docker_container_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_container_id"], m["docker_container_name"], m["domain"], m["docker_image_id"]
	}

	return MappedNode{
		ID:    id,
		Major: major,
		Minor: minor,
		Rank:  rank,
	}, true
}

// ProcessContainerImage maps topology nodes to the container images they run
// on. If no container metadata is found, nodes are grouped into the
// Uncontained node.
func ProcessContainerImage(_ string, m NodeMetadata) (MappedNode, bool) {
	var id, major, minor, rank string
	if m["docker_image_id"] == "" {
		id, major, minor, rank = "uncontained", "Uncontained", "", "uncontained"
	} else {
		id, major, minor, rank = m["docker_image_id"], m["docker_image_name"], "", m["docker_image_id"]
	}

	return MappedNode{
		ID:    id,
		Major: major,
		Minor: minor,
		Rank:  rank,
	}, true
}

// NetworkHostname takes a node NodeMetadata and returns a representation
// based on the hostname. Major label is the hostname, the minor label is the
// domain, if any.
func NetworkHostname(_ string, m NodeMetadata) (MappedNode, bool) {
	var (
		name   = m["name"]
		domain = ""
		parts  = strings.SplitN(name, ".", 2)
	)

	if len(parts) == 2 {
		domain = parts[1]
	}

	return MappedNode{
		ID:    fmt.Sprintf("host:%s", name),
		Major: parts[0],
		Minor: domain,
		Rank:  parts[0],
	}, name != ""
}

// GenericPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericPseudoNode(src string, srcMapped RenderableNode, dst string) (MappedNode, bool) {
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

	return MappedNode{
		ID:    outputID,
		Major: maj,
		Minor: min,
	}, true
}

// GenericGroupedPseudoNode contains heuristics for building sensible pseudo nodes.
// It should go away.
func GenericGroupedPseudoNode(src string, srcMapped RenderableNode, dst string) (MappedNode, bool) {
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

	return MappedNode{
		ID:    outputID,
		Major: maj,
		Minor: min,
	}, true
}

// InternetOnlyPseudoNode never creates a pseudo node, unless it's the Internet.
func InternetOnlyPseudoNode(_ string, _ RenderableNode, dst string) (MappedNode, bool) {
	if dst == TheInternet {
		return MappedNode{ID: TheInternet, Major: humanTheInternet}, true
	}
	return MappedNode{}, false
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
