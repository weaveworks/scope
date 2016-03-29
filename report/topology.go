package report

import (
	"fmt"
	"strings"
)

// Topology describes a specific view of a network. It consists of nodes and
// edges, and metadata about those nodes and edges, represented by
// EdgeMetadatas and Nodes respectively. Edges are directional, and embedded
// in the Node struct.
type Topology struct {
	Nodes    `json:"nodes"`
	Controls `json:"controls,omitempty"`
}

// MakeTopology gives you a Topology.
func MakeTopology() Topology {
	return Topology{
		Nodes:    map[string]Node{},
		Controls: Controls{},
	}
}

// AddNode adds node to the topology under key nodeID; if a
// node already exists for this key, nmd is merged with that node.
// The same topology is returned to enable chaining.
// This method is different from all the other similar methods
// in that it mutates the Topology, to solve issues of GC pressure.
func (t Topology) AddNode(nodeID string, node Node) Topology {
	if existing, ok := t.Nodes[nodeID]; ok {
		node = node.Merge(existing)
	}
	t.Nodes[nodeID] = node
	return t
}

// Copy returns a value copy of the Topology.
func (t Topology) Copy() Topology {
	return Topology{
		Nodes:    t.Nodes.Copy(),
		Controls: t.Controls.Copy(),
	}
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (t Topology) Merge(other Topology) Topology {
	return Topology{
		Nodes:    t.Nodes.Merge(other.Nodes),
		Controls: t.Controls.Merge(other.Controls),
	}
}

// Nodes is a collection of nodes in a topology. Keys are node IDs.
// TODO(pb): type Topology map[string]Node
type Nodes map[string]Node

// Copy returns a value copy of the Nodes.
func (n Nodes) Copy() Nodes {
	cp := make(Nodes, len(n))
	for k, v := range n {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (n Nodes) Merge(other Nodes) Nodes {
	cp := n.Copy()
	for k, v := range other {
		if n, ok := cp[k]; ok { // don't overwrite
			v = v.Merge(n)
		}
		cp[k] = v
	}
	return cp
}

// Prune returns a copy of the Nodes with all information not strictly
// necessary for rendering nodes and edges in the UI cut away.
func (n Nodes) Prune() Nodes {
	result := Nodes{}
	for id, node := range n {
		result[id] = node.Prune()
	}
	return result
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	errs := []string{}

	// Check all nodes are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID, nmd := range t.Nodes {
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}

		// Check all adjancency keys has entries in Node.
		for _, dstNodeID := range nmd.Adjacency {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node missing from adjacency %q -> %q", nodeID, dstNodeID))
			}
		}

		// Check all the edge metadatas have entries in adjacencies
		nmd.Edges.ForEach(func(dstNodeID string, _ EdgeMetadata) {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node %s missing for edge %q", dstNodeID, nodeID))
			}
		})
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}

	return nil
}
