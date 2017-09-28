package report

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/test/reflect"
)

// NodeSet is a set of nodes, as a slice sorted by ID. Clients must use
// the Add method to add nodes
type NodeSet struct {
	entries nodesByID
}

// nodesByID implements sort.Interface.
type nodesByID []*Node

func (m nodesByID) Len() int           { return len(m) }
func (m nodesByID) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m nodesByID) Less(i, j int) bool { return m[i].ID < m[j].ID }

var emptyNodeSet = NodeSet{}

// MakeNodeSet makes a new NodeSet with the given nodes.
func MakeNodeSet(nodes ...Node) NodeSet {
	return emptyNodeSet.Add(nodes...)
}

// Add adds the nodes to the NodeSet. Add is the only valid way to grow a
// NodeSet. Add returns the NodeSet to enable chaining.
func (n NodeSet) Add(nodes ...Node) NodeSet {
	if len(nodes) == 0 {
		return n
	}
	result := make(nodesByID, len(n.entries), len(n.entries)+len(nodes))
	copy(result, n.entries)
	addends := make(nodesByID, len(nodes))
	for i := 0; i < len(nodes); i++ {
		addends[i] = &nodes[i]
	}
	sort.Sort(addends)
	for _, val := range addends {
		i := sort.Search(len(result), func(i int) bool {
			return result[i].ID >= val.ID
		})
		// i is now the position where val should go, either at the end or in the middle
		if i == len(result) {
			result = append(result, nil)
		} else if result[i].ID != val.ID {
			result = append(result, nil)
			copy(result[i+1:], result[i:])
		}
		result[i] = val
	}
	return NodeSet{result}
}

// Delete deletes the nodes from the NodeSet by ID. Delete is the only valid
// way to shrink a NodeSet. Delete returns the NodeSet to enable chaining.
func (n NodeSet) Delete(ids ...string) NodeSet {
	if n.Size() == 0 {
		return n
	}
	result := make(nodesByID, len(n.entries))
	copy(result, n.entries)
	for _, id := range ids {
		i := sort.Search(len(result), func(i int) bool {
			return result[i].ID >= id
		})
		if i < len(result) && result[i].ID == id {
			copy(result[i:], result[i+1:])
			result = result[:len(result)-1]
		}
	}
	return NodeSet{result}
}

// Merge combines the two NodeSets and returns a new result.
func (m NodeSet) Merge(n NodeSet) NodeSet {
	switch {
	case m.entries == nil:
		return n
	case n.entries == nil:
		return m
	}
	out := make(nodesByID, 0, len(m.entries)+len(n.entries))

	i, j := 0, 0
	for i < len(m.entries) {
		switch {
		case j >= len(n.entries) || m.entries[i].ID < n.entries[j].ID:
			out = append(out, m.entries[i])
			i++
		case m.entries[i].ID == n.entries[j].ID:
			i++
			fallthrough
		default:
			out = append(out, n.entries[j])
			j++
		}
	}
	for ; j < len(n.entries); j++ {
		out = append(out, n.entries[j])
	}
	return NodeSet{out}
}

// Lookup the node 'key'
func (n NodeSet) Lookup(key string) (Node, bool) {
	i := sort.Search(len(n.entries), func(i int) bool {
		return n.entries[i].ID >= key
	})
	if i < len(n.entries) && n.entries[i].ID == key {
		return *n.entries[i], true
	}
	return Node{}, false
}

// Size is the number of nodes in the set
func (n NodeSet) Size() int {
	return len(n.entries)
}

// ForEach executes f for each node in the set.
func (n NodeSet) ForEach(f func(Node)) {
	for _, value := range n.entries {
		f(*value)
	}
}

func (n NodeSet) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range n.entries {
		fmt.Fprintf(buf, "%s: %s, ", val.ID, spew.Sdump(val))
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other NodeSets
func (n NodeSet) DeepEqual(o NodeSet) bool {
	if n.Size() != o.Size() {
		return false
	}
	for i := range n.entries {
		if !reflect.DeepEqual(n.entries[i], o.entries[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer
func (n *NodeSet) CodecEncodeSelf(encoder *codec.Encoder) {
	encoder.Encode(n.entries)
}

// CodecDecodeSelf implements codec.Selfer
func (n *NodeSet) CodecDecodeSelf(decoder *codec.Decoder) {
	in := nodesByID{}
	if err := decoder.Decode(&in); err != nil {
		return
	}
	*n = NodeSet{in}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (NodeSet) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*NodeSet) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
