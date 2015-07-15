package render

import (
	"github.com/weaveworks/scope/report"
)

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID         string        `json:"id"`                    //
	LabelMajor string        `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor string        `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank       string        `json:"rank"`                  // to help the layout engine
	Pseudo     bool          `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Adjacency  report.IDList `json:"adjacency,omitempty"`   // Node IDs (in the same topology domain)
	Origins    report.IDList `json:"origins,omitempty"`     // Core node IDs that contributed information

	AggregateMetadata   `json:"metadata"` // Numeric sums
	report.NodeMetadata `json:"-"`        // merged NodeMetadata of the nodes used to build this
}

// RenderableNodes is a set of RenderableNodes
type RenderableNodes map[string]RenderableNode

// Merge merges two sets of RenderableNodes
func (rns RenderableNodes) Merge(other RenderableNodes) {
	for key, value := range other {
		if existing, ok := rns[key]; ok {
			existing.Merge(value)
			rns[key] = existing
		} else {
			rns[key] = value
		}
	}
}

// Merge merges in another RenderableNode
func (rn *RenderableNode) Merge(other RenderableNode) {
	if rn.LabelMajor == "" {
		rn.LabelMajor = other.LabelMajor
	}

	if rn.LabelMinor == "" {
		rn.LabelMinor = other.LabelMinor
	}

	if rn.Rank == "" {
		rn.Rank = other.Rank
	}

	if rn.Pseudo != other.Pseudo {
		panic(rn.ID)
	}

	rn.Adjacency = rn.Adjacency.Merge(other.Adjacency)
	rn.Origins = rn.Origins.Merge(other.Origins)

	rn.AggregateMetadata.Merge(other.AggregateMetadata)
	rn.NodeMetadata.Merge(other.NodeMetadata)
}

// NewRenderableNode makes a new RenderableNode
func NewRenderableNode(id, major, minor, rank string, nmd report.NodeMetadata) RenderableNode {
	return RenderableNode{
		ID:                id,
		LabelMajor:        major,
		LabelMinor:        minor,
		Rank:              rank,
		Pseudo:            false,
		AggregateMetadata: AggregateMetadata{},
		NodeMetadata:      nmd.Copy(),
	}
}

func newDerivedNode(id string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:                id,
		LabelMajor:        "",
		LabelMinor:        "",
		Rank:              "",
		Pseudo:            node.Pseudo,
		AggregateMetadata: node.AggregateMetadata,
		Origins:           node.Origins,
		NodeMetadata:      report.NewNodeMetadata(report.Metadata{}),
	}
}

func newPseudoNode(id, major, minor string) RenderableNode {
	return RenderableNode{
		ID:                id,
		LabelMajor:        major,
		LabelMinor:        minor,
		Rank:              "",
		Pseudo:            true,
		AggregateMetadata: AggregateMetadata{},
		NodeMetadata:      report.NewNodeMetadata(report.Metadata{}),
	}
}

func newDerivedPseudoNode(id, major string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:                id,
		LabelMajor:        major,
		LabelMinor:        "",
		Rank:              "",
		Pseudo:            true,
		AggregateMetadata: node.AggregateMetadata,
		Origins:           node.Origins,
		NodeMetadata:      report.NewNodeMetadata(report.Metadata{}),
	}
}
