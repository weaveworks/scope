package render

import (
	"github.com/weaveworks/scope/report"
)

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID         string                   `json:"id"`                    //
	LabelMajor string                   `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor string                   `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank       string                   `json:"rank"`                  // to help the layout engine
	Pseudo     bool                     `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Adjacency  report.IDList            `json:"adjacency,omitempty"`   // Node IDs (in the same topology domain)
	Origins    report.IDList            `json:"origins,omitempty"`     // Core node IDs that contributed information
	Metadata   report.AggregateMetadata `json:"metadata"`              // Numeric sums
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

	rn.Adjacency = rn.Adjacency.Add(other.Adjacency...)
	rn.Origins = rn.Origins.Add(other.Origins...)

	rn.Metadata.Merge(other.Metadata)
}
