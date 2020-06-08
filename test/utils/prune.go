package utils

import (
	"strings"
	"time"

	"github.com/weaveworks/scope/report"
)

// Prune returns a copy of the Nodes with all information not strictly
// necessary for rendering nodes and edges in the UI cut away.
func Prune(nodes report.Nodes) report.Nodes {
	result := report.Nodes{}
	for id, node := range nodes {
		result[id] = PruneNode(node)
	}
	return result
}

// PruneNode returns a copy of the Node with all information not strictly
// necessary for rendering nodes and edges stripped away. Specifically, that
// means cutting out parts of the Node.
func PruneNode(node report.Node) report.Node {
	prunedNode := report.MakeNode(
		node.ID).
		WithTopology(node.Topology).
		WithAdjacent(node.Adjacency...).
		WithChildren(node.ChildIDs)
	// Copy counters across, but with zero timestamp so they compare equal
	node.Latest.ForEach(func(k string, _ time.Time, v string) {
		if strings.HasPrefix(k, report.CounterPrefix) {
			prunedNode.Latest = prunedNode.Latest.Set(k, time.Time{}, v)
		}
	})
	return prunedNode
}
