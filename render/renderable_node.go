package render

import (
	"github.com/weaveworks/scope/report"
)

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID          string            `json:"id"`                    //
	LabelMajor  string            `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor  string            `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank        string            `json:"rank"`                  // to help the layout engine
	Pseudo      bool              `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Children    RenderableNodeSet `json:"children,omitempty"`    // Nodes which have been grouped into this one
	ControlNode string            `json:"control_node"`          // ID of node from which to show the controls in the UI
	Shape       string            `json:"shape"`                 // Shape node should be rendered as
	Stack       bool              `json:"stack"`                 // Should UI render this node as a stack?

	report.EdgeMetadata `json:"metadata"` // Numeric sums
	report.Node
}

// Shapes that are allowed
const (
	Circle   = "circle"
	Square   = "square"
	Heptagon = "heptagon"
	Hexagon  = "hexagon"
	Cloud    = "cloud"
)

// NewRenderableNode makes a new RenderableNode
func NewRenderableNode(id string) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   "",
		LabelMinor:   "",
		Rank:         "",
		Pseudo:       false,
		EdgeMetadata: report.EdgeMetadata{},
		Node:         report.MakeNode(),
		Shape:        Circle,
	}
}

// NewRenderableNodeWith makes a new RenderableNode with some fields filled in
func NewRenderableNodeWith(id, major, minor, rank string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   major,
		LabelMinor:   minor,
		Rank:         rank,
		Pseudo:       false,
		Children:     node.Children.Copy(),
		EdgeMetadata: node.EdgeMetadata.Copy(),
		Node:         node.Node.Copy(),
		Shape:        Circle,
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
		Children:     node.Children.Copy(),
		EdgeMetadata: node.EdgeMetadata.Copy(),
		Node:         node.Node.Copy(),
		ControlNode:  "", // Do not propagate ControlNode when making a derived node!
		Shape:        Circle,
	}
}

func newDerivedPseudoNode(id, major string, node RenderableNode) RenderableNode {
	return RenderableNode{
		ID:           id,
		LabelMajor:   major,
		LabelMinor:   "",
		Rank:         "",
		Pseudo:       true,
		Children:     node.Children.Copy(),
		EdgeMetadata: node.EdgeMetadata.Copy(),
		Node:         node.Node.Copy(),
		Shape:        Circle,
	}
}

// WithNode creates a new RenderableNode based on rn, with n
func (rn RenderableNode) WithNode(n report.Node) RenderableNode {
	result := rn.Copy()
	result.Node = result.Node.Merge(n)
	return result
}

// WithParents creates a new RenderableNode based on rn, where n has the given parents set
func (rn RenderableNode) WithParents(p report.Sets) RenderableNode {
	result := rn.Copy()
	result.Node.Parents = p
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

	if result.ControlNode == "" {
		result.ControlNode = other.ControlNode
	}

	if result.Pseudo != other.Pseudo {
		panic(result.ID)
	}

	result.Stack = result.Stack || rn.Stack
	result.Children = rn.Children.Merge(other.Children)
	result.EdgeMetadata = rn.EdgeMetadata.Merge(other.EdgeMetadata)
	result.Node = rn.Node.Merge(other.Node)

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
		Children:     rn.Children.Copy(),
		EdgeMetadata: rn.EdgeMetadata.Copy(),
		Node:         rn.Node.Copy(),
		ControlNode:  rn.ControlNode,
		Shape:        rn.Shape,
		Stack:        rn.Stack,
	}
}

// Prune returns a copy of the RenderableNode with all information not
// strictly necessary for rendering nodes and edges stripped away.
// Specifically, that means cutting out parts of the Node.
func (rn RenderableNode) Prune() RenderableNode {
	cp := rn.Copy()
	cp.Node = report.MakeNode().WithAdjacent(cp.Node.Adjacency...)
	cp.Children = MakeRenderableNodeSet()
	rn.Children.ForEach(func(n RenderableNode) {
		cp.Children = cp.Children.Add(n.Prune())
	})
	return cp
}

// RenderableNodes is a set of RenderableNodes
type RenderableNodes map[string]RenderableNode

// Copy produces a deep copy of the RenderableNodes
func (rns RenderableNodes) Copy() RenderableNodes {
	result := RenderableNodes{}
	for key, value := range rns {
		result[key] = value.Copy()
	}
	return result
}

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

// Prune returns a copy of the RenderableNodes with all information not
// strictly necessary for rendering nodes and edges in the UI cut away.
func (rns RenderableNodes) Prune() RenderableNodes {
	cp := rns.Copy()
	for id, rn := range cp {
		cp[id] = rn.Prune()
	}
	return cp
}
