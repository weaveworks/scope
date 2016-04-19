package detailed

import (
	"github.com/weaveworks/scope/report"
)

// NodeTables produces a list of tables (to be consumed directly by the UI) based
// on the report and the node.  It uses the report to get the templates for the node's
// topology.
func NodeTables(r report.Report, n report.Node) []report.Table {
	if _, ok := n.Counters.Lookup(n.Topology); ok {
		// This is a group of nodes, so no tables!
		return nil
	}

	if topology, ok := r.Topology(n.Topology); ok {
		return topology.TableTemplates.Tables(n)
	}
	return nil
}
