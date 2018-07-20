package report

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/weaveworks/ps"

	"github.com/weaveworks/scope/test/reflect"
)

// NodeSet is a set of nodes keyed on ID. Clients must use
// the Add method to add nodes
type NodeSet struct {
	psMap ps.Map
}

var emptyNodeSet = NodeSet{ps.NewMap()}

// MakeNodeSet makes a new NodeSet with the given nodes.
func MakeNodeSet(nodes ...Node) NodeSet {
	return emptyNodeSet.Add(nodes...)
}

// Copy returns a value copy of the Nodes.
func (n NodeSet) Copy() NodeSet {
	result := ps.NewMap()
	n.ForEach(func(node Node) {
		result = result.UnsafeMutableSet(node.ID, node)
	})
	return NodeSet{result}
}

// UnsafeAdd adds a node to the NodeSet. Only call this if n has one owner.
func (n *NodeSet) UnsafeAdd(node Node) {
	if n.psMap == nil {
		n.psMap = ps.NewMap()
	}
	n.psMap = n.psMap.UnsafeMutableSet(node.ID, node)
}

// Add adds the nodes to the NodeSet, and returns the NodeSet to enable chaining.
func (n NodeSet) Add(nodes ...Node) NodeSet {
	if len(nodes) == 0 {
		return n
	}
	result := n.psMap
	if result == nil {
		result = ps.NewMap()
	}
	for _, node := range nodes {
		result = result.Set(node.ID, node)
	}
	return NodeSet{result}
}

// Delete deletes the nodes from the NodeSet by ID. Delete is the only valid
// way to shrink a NodeSet. Delete returns the NodeSet to enable chaining.
func (n NodeSet) Delete(ids ...string) NodeSet {
	if n.Size() == 0 {
		return n
	}
	result := n.psMap
	for _, id := range ids {
		result = result.Delete(id)
	}
	if result.Size() == 0 {
		return emptyNodeSet
	}
	return NodeSet{result}
}

// UnsafeMerge combines the two NodeSets, altering n
func (n *NodeSet) UnsafeMerge(other NodeSet) {
	if other.psMap == nil || other.psMap.Size() == 0 {
		return
	}
	if n.psMap == nil {
		n.psMap = ps.NewMap()
	}
	other.psMap.ForEach(func(key string, otherVal interface{}) {
		n.psMap = n.psMap.UnsafeMutableSet(key, otherVal)
	})
}

// Merge combines the two NodeSets and returns a new result.
func (n NodeSet) Merge(other NodeSet) NodeSet {
	nSize, otherSize := n.Size(), other.Size()
	if nSize == 0 {
		return other
	}
	if otherSize == 0 {
		return n
	}
	result, iter := n.psMap, other.psMap
	if nSize < otherSize {
		result, iter = iter, result
	}
	iter.ForEach(func(key string, otherVal interface{}) {
		result = result.Set(key, otherVal)
	})
	return NodeSet{result}
}

// Lookup the node 'key'
func (n NodeSet) Lookup(key string) (Node, bool) {
	if n.psMap != nil {
		value, ok := n.psMap.Lookup(key)
		if ok {
			return value.(Node), true
		}
	}
	return Node{}, false
}

// Size is the number of nodes in the set
func (n NodeSet) Size() int {
	if n.psMap == nil {
		return 0
	}
	return n.psMap.Size()
}

// ForEach executes f for each node in the set.
func (n NodeSet) ForEach(f func(Node)) {
	if n.psMap != nil {
		n.psMap.ForEach(func(_ string, val interface{}) {
			f(val.(Node))
		})
	}
}

func (n NodeSet) String() string {
	buf := bytes.NewBufferString("{")
	for _, key := range mapKeys(n.psMap) {
		val, _ := n.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %s, ", key, spew.Sdump(val))
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other NodeSets
func (n NodeSet) DeepEqual(o NodeSet) bool {
	return mapEqual(n.psMap, o.psMap, reflect.DeepEqual)
}
