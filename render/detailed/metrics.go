package detailed

import (
	"github.com/weaveworks/scope/report"
)

// NodeMetrics produces a table (to be consumed directly by the UI) based on
// an a report.Node, which is (hopefully) a node in one of our topologies.
func NodeMetrics(r report.Report, n report.Node) []report.MetricRow {
	if _, ok := n.Counters.Lookup(n.Topology); ok {
		// This is a group of nodes, so no metrics!
		return nil
	}

	topology, ok := r.Topology(n.Topology)
	if !ok {
		return nil
	}
	return topology.MetricTemplates.MetricRows(n)
}
