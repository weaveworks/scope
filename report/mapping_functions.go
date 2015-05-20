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
func ProcessPID(id string, m NodeMetadata, grouped bool) (MappedNode, bool) {
	var (
		domain = m["domain"]
		pid    = m["pid"]
		name   = m["name"]
		minor  = fmt.Sprintf("%s (%s)", domain, pid)
	)

	if grouped {
		domain = ""
		minor = ""
	}

	return MappedNode{
		ID:    fmt.Sprintf("pid:%s:%s", domain, pid),
		Major: name,
		Minor: minor,
		Rank:  pid,
	}, pid != ""
}

// ProcessCgroup takes a node NodeMetadata from a Process topology, augmented
// with cgroup fields, and returns a representation based on the cgroup. If
// the cgroup is not present, it falls back to process name.
func ProcessCgroup(id string, m NodeMetadata, grouped bool) (MappedNode, bool) {
	var (
		domain = m["domain"]
		cgroup = m["cgroup"]
	)

	if cgroup == "" {
		cgroup = m["name"]
	}

	if grouped {
		domain = ""
	}

	return MappedNode{
		ID:    fmt.Sprintf("cgroup:%s:%s", domain, cgroup),
		Major: cgroup,
		Minor: domain,
		Rank:  cgroup,
	}, cgroup != ""
}

// ProcessName takes a node NodeMetadata from a Process topology, and returns
// a representation based on the process name.
func ProcessName(id string, m NodeMetadata, grouped bool) (MappedNode, bool) {
	var (
		name   = m["name"]
		domain = m["domain"]
	)

	if grouped {
		domain = ""
	}

	return MappedNode{
		ID:    fmt.Sprintf("proc:%s:%s", domain, name),
		Major: name,
		Minor: domain,
		Rank:  name,
	}, name != ""
}

// NetworkHostname takes a node NodeMetadata from a Network topology, and
// returns a representation based on the hostname. Major label is the
// hostname, the minor label is the domain, if any.
func NetworkHostname(id string, m NodeMetadata, _ bool) (MappedNode, bool) {
	var (
		name   = m["name"]
		domain = ""
		parts  = strings.SplitN(name, ".", 2)
	)

	if len(parts) == 2 {
		domain = parts[1]
	}

	// Note: no grouped special case.

	return MappedNode{
		ID:    fmt.Sprintf("host:%s", name),
		Major: parts[0],
		Minor: domain,
		Rank:  parts[0],
	}, name != ""
}

// NetworkIP takes a node NodeMetadata from a Network topology, and returns a
// representation based on the (scoped) IP. Major label is the IP, the Minor
// label is the hostname.
func NetworkIP(id string, m NodeMetadata, _ bool) (MappedNode, bool) {
	var (
		name = m["name"]
		ip   = strings.SplitN(id, ScopeDelim, 2)[1]
	)

	// Note: no grouped special case.

	return MappedNode{
		ID:    fmt.Sprintf("addr:%s", id),
		Major: ip,
		Minor: name,
		Rank:  ip,
	}, id != ""
}

// MapFuncRegistry maps a string to a MapFunc.
var MapFuncRegistry = map[string]MapFunc{
	"processpid":      ProcessPID,
	"processcgroup":   ProcessCgroup,
	"processname":     ProcessName,
	"networkhostname": NetworkHostname,
	"networkip":       NetworkIP,
}
