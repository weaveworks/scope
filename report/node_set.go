package report

import (
	"sort"
)

// NodeSet is a sorted set of nodes keyed on (Topology, ID). Clients must use
// the Add method to add nodes
type NodeSet []Node

// MakeNodeSet makes a new NodeSet with the given nodes.
func MakeNodeSet(nodes ...Node) NodeSet {
	if len(nodes) <= 0 {
		return nil
	}
	result := make(NodeSet, len(nodes))
	copy(result, nodes)
	sort.Sort(result)
	for i := 1; i < len(result); { // remove any duplicates
		if result[i-1].Equal(result[i]) {
			result = append(result[:i-1], result[i:]...)
			continue
		}
		i++
	}
	return result
}

// Implementation of sort.Interface
func (n NodeSet) Len() int           { return len(n) }
func (n NodeSet) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n NodeSet) Less(i, j int) bool { return n[i].Before(n[j]) }

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
		n[i] = node
	}
	return n
}

// Merge combines the two NodeSets and returns a new result.
func (n NodeSet) Merge(other NodeSet) NodeSet {
	switch {
	case len(other) <= 0: // Optimise special case, to avoid allocating
		return n // (note unit test DeepEquals breaks if we don't do this)
	case len(n) <= 0:
		return other
	}

	result := make([]Node, 0, len(n)+len(other))
	for len(n) > 0 || len(other) > 0 {
		switch {
		case len(n) == 0:
			return append(result, other...)
		case len(other) == 0:
			return append(result, n...)
		case n[0].Before(other[0]):
			result = append(result, n[0])
			n = n[1:]
		case n[0].After(other[0]):
			result = append(result, other[0])
			other = other[1:]
		default: // equal
			result = append(result, other[0])
			n = n[1:]
			other = other[1:]
		}
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
		result[i] = node
	}
	return result
}
