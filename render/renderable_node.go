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
	Origins    report.IDList `json:"origins,omitempty"`     // Core node IDs that contributed information

	report.EdgeMetadata `json:"metadata"` // Numeric sums
	report.NodeMetadata
}

// NewRenderableNode makes a new RenderableNode
func NewRenderableNode(id string) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   "",
		LabelMinor:   "",
		Rank:         "",
		Pseudo:       false,
		Origins:      report.MakeIDList(),
		EdgeMetadata: report.EdgeMetadata{},
		NodeMetadata: report.MakeNodeMetadata(),
	}
}

// NewRenderableNodeWith makes a new RenderableNode with some fields filled in
func NewRenderableNodeWith(id, major, minor, rank string, rn RenderableNode) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   major,
		LabelMinor:   minor,
		Rank:         rank,
		Pseudo:       false,
		Origins:      rn.Origins.Copy(),
		EdgeMetadata: rn.EdgeMetadata.Copy(),
		NodeMetadata: rn.NodeMetadata.Copy(),
	}
}

// NewDerivedNode create a renderable node based on node, but with a new ID
func NewDerivedNode(id string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   "",
		LabelMinor:   "",
		Rank:         "",
		Pseudo:       node.Pseudo,
		Origins:      node.Origins.Copy(),
		EdgeMetadata: node.EdgeMetadata.Copy(),
		NodeMetadata: node.NodeMetadata.Copy(),
	}
}

func newDerivedPseudoNode(id, major string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   major,
		LabelMinor:   "",
		Rank:         "",
		Pseudo:       true,
		Origins:      node.Origins.Copy(),
		EdgeMetadata: node.EdgeMetadata.Copy(),
		NodeMetadata: node.NodeMetadata.Copy(),
	}
}

// WithNodeMetadata creates a new RenderableNode based on rn, with n
func (rn RenderableNode) WithNodeMetadata(n report.NodeMetadata) RenderableNode {
	result := rn.Copy()
	result.NodeMetadata = result.NodeMetadata.Merge(n)
	return result
}

// Merge merges rn with other and returns a new RenderableNode
func (rn RenderableNode) Merge(other RenderableNode) RenderableNode {
	result := rn.Copy()

	if result.LabelMajor == "" {
		result.LabelMajor = other.LabelMajor
	}

	if result.LabelMinor == "" {
		result.LabelMinor = other.LabelMinor
	}

	if result.Rank == "" {
		result.Rank = other.Rank
	}

	if result.Pseudo != other.Pseudo {
		panic(result.ID)
	}

	result.Origins = rn.Origins.Merge(other.Origins)
	result.EdgeMetadata = rn.EdgeMetadata.Merge(other.EdgeMetadata)
	result.NodeMetadata = rn.NodeMetadata.Merge(other.NodeMetadata)

	return result
}

// Copy makes a deep copy of rn
func (rn RenderableNode) Copy() RenderableNode {
	return RenderableNode{
		ID:           rn.ID,
		LabelMajor:   rn.LabelMajor,
		LabelMinor:   rn.LabelMinor,
		Rank:         rn.Rank,
		Pseudo:       rn.Pseudo,
		Origins:      rn.Origins.Copy(),
		EdgeMetadata: rn.EdgeMetadata.Copy(),
		NodeMetadata: rn.NodeMetadata.Copy(),
	}
}

// RenderableNodes is a set of RenderableNodes
type RenderableNodes map[string]RenderableNode

// Merge merges two sets of RenderableNodes, returning a new set.
func (rns RenderableNodes) Merge(other RenderableNodes) RenderableNodes {
	result := RenderableNodes{}
	for key, value := range rns {
		result[key] = value
	}
	for key, value := range other {
		existing, ok := result[key]
		if ok {
			value = value.Merge(existing)
		}
		result[key] = value
	}
	return result
}
