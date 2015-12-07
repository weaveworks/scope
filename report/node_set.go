package report

import (
	"sort"
)

// NodeSet is a sorted set of nodes keyed on (Topology, ID). Clients must use
// the Add method to add nodes
type NodeSet []Node

// MakeNodeSet makes a new NodeSet with the given nodes.
// TODO: Make this more efficient
func MakeNodeSet(nodes ...Node) NodeSet {
	if len(nodes) <= 0 {
		return nil
	}
	result := NodeSet{}
	for _, node := range nodes {
		result = result.Add(node)
	}
	return result
}

// Add adds the nodes to the NodeSet. Add is the only valid way to grow a
// NodeSet. Add returns the NodeSet to enable chaining.
func (n NodeSet) Add(nodes ...Node) NodeSet {
	for _, node := range nodes {
		i := sort.Search(len(n), func(i int) bool {
			return n[i].Topology >= node.Topology && n[i].ID >= node.ID
		})
		if i < len(n) && n[i].Topology == node.Topology && n[i].ID == node.ID {
			// The list already has the element.
			continue
		}
		// It a new element, insert it in order.
		n = append(n, Node{})
		copy(n[i+1:], n[i:])
		n[i] = node.Copy()
	}
	return n
}

// Merge combines the two NodeSets and returns a new result.
// TODO: Make this more efficient
func (n NodeSet) Merge(other NodeSet) NodeSet {
	switch {
	case len(other) <= 0: // Optimise special case, to avoid allocating
		return n // (note unit test DeepEquals breaks if we don't do this)
	case len(n) <= 0:
		return other
	}
	result := n.Copy()
	for _, node := range other {
		result = result.Add(node)
	}
	return result
}

// Copy returns a value copy of the NodeSet.
func (n NodeSet) Copy() NodeSet {
	if n == nil {
		return n
	}
	result := make(NodeSet, len(n))
	for i, node := range n {
		result[i] = node.Copy()
	}
	return result
}
