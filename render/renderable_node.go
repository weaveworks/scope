package render

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mndrix/ps"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID          string         `json:"id"`                    //
	LabelMajor  string         `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor  string         `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank        string         `json:"rank"`                  // to help the layout engine
	Pseudo      bool           `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Children    report.NodeSet `json:"children,omitempty"`    // Nodes which have been grouped into this one
	ControlNode string         `json:"-"`                     // ID of node from which to show the controls in the UI

	report.EdgeMetadata `json:"metadata"` // Numeric sums
	report.Node
}

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
// Note: This is non-commutative, due to ID and Topology fields.
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
	}
}

// Prune returns a copy of the RenderableNode with all information not
// strictly necessary for rendering nodes and edges stripped away.
// Specifically, that means cutting out parts of the Node.
func (rn RenderableNode) Prune() RenderableNode {
	cp := rn.Copy()
	cp.Node = report.MakeNode().WithAdjacent(cp.Node.Adjacency...)
	cp.Children = nil
	return cp
}

// RenderableNodes is a set of RenderableNodes
type RenderableNodes struct {
	psMap ps.Map
}

// EmptyRenderableNodes is an empty set of renderable nodes
var EmptyRenderableNodes = RenderableNodes{ps.NewMap()}

// MakeRenderableNodes makes a set with the given RenderableNodes
func MakeRenderableNodes(nodes ...RenderableNode) RenderableNodes {
	return EmptyRenderableNodes.Add(nodes...)
}

// Add adds each node into the set. Unlike merge, this overwrites existing
// nodes with the same ID.
func (rns RenderableNodes) Add(nodes ...RenderableNode) RenderableNodes {
	if rns.psMap == nil {
		rns = EmptyRenderableNodes
	}
	psMap := rns.psMap
	for _, node := range nodes {
		psMap = psMap.Set(node.ID, node)
	}
	return RenderableNodes{psMap}
}

// Delete removes the node with id
func (rns RenderableNodes) Delete(id string) RenderableNodes {
	if rns.psMap == nil {
		rns = EmptyRenderableNodes
	}
	return RenderableNodes{rns.psMap.Delete(id)}
}

// Copy is a noop
func (rns RenderableNodes) Copy() RenderableNodes {
	return rns
}

func (rns RenderableNodes) String() string {
	if rns.psMap == nil {
		return "{}"
	}
	buf := bytes.NewBufferString("{")
	for _, key := range rns.Keys() {
		val, _ := rns.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %v, ", key, val)
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// Lookup looks up a renderable node by id
func (rns RenderableNodes) Lookup(id string) (RenderableNode, bool) {
	if rns.psMap == nil {
		return RenderableNode{}, false
	}
	if val, ok := rns.psMap.Lookup(id); ok {
		return val.(RenderableNode), true
	}
	return RenderableNode{}, false
}

// Keys returns the keys present in this set, in sorted order
func (rns RenderableNodes) Keys() []string {
	if rns.psMap == nil {
		return nil
	}
	keys := rns.psMap.Keys()
	sort.Strings(keys)
	return keys
}

// Size returns the number of nodes
func (rns RenderableNodes) Size() int {
	if rns.psMap == nil {
		return 0
	}
	return rns.psMap.Size()
}

// ForEach executes f for each node
func (rns RenderableNodes) ForEach(f func(RenderableNode)) {
	if rns.psMap != nil {
		rns.psMap.ForEach(func(key string, val interface{}) {
			f(val.(RenderableNode))
		})
	}
}

// Merge merges two sets of RenderableNodes, returning a new set.
// Note: Because merging RenderableNodes is non-commutative, neither is this.
func (rns RenderableNodes) Merge(other RenderableNodes) RenderableNodes {
	switch {
	case rns.Size() == 0:
		return other
	case other.Size() == 0:
		return rns
	}
	result := rns.psMap
	other.ForEach(func(node RenderableNode) {
		if existing, ok := result.Lookup(node.ID); ok {
			node = node.Merge(existing.(RenderableNode))
		}
		result = result.Set(node.ID, node)
	})

	return RenderableNodes{result}
}

// Prune returns a copy of the RenderableNodes with all information not
// strictly necessary for rendering nodes and edges in the UI cut away.
func (rns RenderableNodes) Prune() RenderableNodes {
	result := ps.NewMap()
	rns.ForEach(func(node RenderableNode) {
		result = result.Set(node.ID, node.Prune())
	})
	return RenderableNodes{result}
}

// DeepEqual tests equality with other RenderableNodes
func (rns RenderableNodes) DeepEqual(d RenderableNodes) bool {
	if rns.Size() != d.Size() {
		return false
	}
	if rns.Size() == 0 {
		return true
	}

	equal := true
	rns.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := d.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

func (rns RenderableNodes) toIntermediate() map[string]RenderableNode {
	intermediate := map[string]RenderableNode{}
	rns.ForEach(func(node RenderableNode) {
		intermediate[node.ID] = node
	})
	return intermediate
}

func (rns RenderableNodes) fromIntermediate(in map[string]RenderableNode) RenderableNodes {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return RenderableNodes{out}
}

// MarshalJSON implements json.Marshaller
func (rns RenderableNodes) MarshalJSON() ([]byte, error) {
	if rns.psMap != nil {
		return json.Marshal(rns.toIntermediate())
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler
func (rns *RenderableNodes) UnmarshalJSON(input []byte) error {
	in := map[string]RenderableNode{}
	if err := json.Unmarshal(input, &in); err != nil {
		return err
	}
	*rns = RenderableNodes{}.fromIntermediate(in)
	return nil
}

// GobEncode implements gob.Marshaller
func (rns RenderableNodes) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(rns.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (rns *RenderableNodes) GobDecode(input []byte) error {
	in := map[string]RenderableNode{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*rns = RenderableNodes{}.fromIntermediate(in)
	return nil
}
