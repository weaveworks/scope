package detailed

import (
	"github.com/weaveworks/scope/report"
)

// NodePropertyLists produces a list of property lists (to be consumed directly by the UI) based
// on the report and the node.  It uses the report to get the templates for the node's topology.
func NodePropertyLists(r report.Report, n report.Node) []report.PropertyList {
	if _, ok := n.Counters.Lookup(n.Topology); ok {
		// This is a group of nodes, so no tables!
		return nil
	}

	if topology, ok := r.Topology(n.Topology); ok {
		return topology.PropertyListTemplates.PropertyLists(n)
	}
	return nil
}
