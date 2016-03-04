package detailed

import (
	"github.com/weaveworks/scope/report"
)

// NodeMetadata produces a table (to be consumed directly by the UI) based on
// an a report.Node, which is (hopefully) a node in one of our topologies.
func NodeMetadata(r report.Report, n report.Node) []report.MetadataRow {
	if _, ok := n.Counters.Lookup(n.Topology); ok {
		// This is a group of nodes, so no metadata!
		return nil
	}

	if topology, ok := r.Topology(n.Topology); ok {
		return topology.MetadataTemplates.MetadataRows(n)
	}
	return nil
}
