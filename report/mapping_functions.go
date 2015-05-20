package report

import (
	"fmt"
	"strings"
)

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
type MapFunc func(string, NodeMetadata, bool) (MappedNode, bool)

// ProcessPID takes a node NodeMetadata from a Process topology, and returns a
// representation with the ID based on the process PID and the labels based
// on the process name.
func ProcessPID(_ string, m NodeMetadata, grouped bool) (MappedNode, bool) {
	var (
		identifier = fmt.Sprintf("%s:%s:%s", "pid", m["domain"], m["pid"])
		minor      = fmt.Sprintf("%s (%s)", m["domain"], m["pid"])
		show       = m["pid"] != "" && m["name"] != ""
	)

	if grouped {
		identifier = m["name"] // flatten
		minor = ""             // nothing meaningful to put here?
	}

	return MappedNode{
		ID:    identifier,
		Major: m["name"],
		Minor: minor,
		Rank:  m["pid"],
	}, show
}

// NetworkHostname takes a node NodeMetadata from a Network topology, and
// returns a representation based on the hostname. Major label is the
// hostname, the minor label is the domain, if any.
func NetworkHostname(_ string, m NodeMetadata, _ bool) (MappedNode, bool) {
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
