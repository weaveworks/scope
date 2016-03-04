package render

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/mndrix/ps"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/test/reflect"
)

// RenderableNodeSet is a set of nodes keyed on (Topology, ID). Clients must use
// the Add method to add nodes
type RenderableNodeSet struct {
	psMap ps.Map
}

// EmptyRenderableNodeSet is the empty set of nodes.
var EmptyRenderableNodeSet = RenderableNodeSet{ps.NewMap()}

// MakeRenderableNodeSet makes a new RenderableNodeSet with the given nodes.
func MakeRenderableNodeSet(nodes ...RenderableNode) RenderableNodeSet {
	return EmptyRenderableNodeSet.Add(nodes...)
}

// Add adds the nodes to the RenderableNodeSet. Add is the only valid way to grow a
// RenderableNodeSet. Add returns the RenderableNodeSet to enable chaining.
func (n RenderableNodeSet) Add(nodes ...RenderableNode) RenderableNodeSet {
	result := n.psMap
	if result == nil {
		result = ps.NewMap()
	}
	for _, node := range nodes {
		result = result.Set(fmt.Sprintf("%s|%s", node.Topology, node.ID), node)
	}
	return RenderableNodeSet{result}
}

// Merge combines the two RenderableNodeSets and returns a new result.
func (n RenderableNodeSet) Merge(other RenderableNodeSet) RenderableNodeSet {
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
	return RenderableNodeSet{result}
}

// Lookup the node 'key'
func (n RenderableNodeSet) Lookup(key string) (RenderableNode, bool) {
	if n.psMap != nil {
		value, ok := n.psMap.Lookup(key)
		if ok {
			return value.(RenderableNode), true
		}
	}
	return RenderableNode{}, false
}

// Keys is a list of all the keys in this set.
func (n RenderableNodeSet) Keys() []string {
	if n.psMap == nil {
		return nil
	}
	k := n.psMap.Keys()
	sort.Strings(k)
	return k
}

// Size is the number of nodes in the set
func (n RenderableNodeSet) Size() int {
	if n.psMap == nil {
		return 0
	}
	return n.psMap.Size()
}

// ForEach executes f for each node in the set. Nodes are traversed in sorted
// order.
func (n RenderableNodeSet) ForEach(f func(RenderableNode)) {
	for _, key := range n.Keys() {
		if val, ok := n.psMap.Lookup(key); ok {
			f(val.(RenderableNode))
		}
	}
}

// Copy is a noop
func (n RenderableNodeSet) Copy() RenderableNodeSet {
	return n
}

func (n RenderableNodeSet) String() string {
	keys := []string{}
	if n.psMap == nil {
		n = EmptyRenderableNodeSet
	}
	psMap := n.psMap
	if psMap == nil {
		psMap = ps.NewMap()
	}
	for _, k := range psMap.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %s, ", key, spew.Sdump(val))
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other RenderableNodeSets
func (n RenderableNodeSet) DeepEqual(i interface{}) bool {
	d, ok := i.(RenderableNodeSet)
	if !ok {
		return false
	}

	if n.Size() != d.Size() {
		return false
	}
	if n.Size() == 0 {
		return true
	}

	equal := true
	n.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := d.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

func (n RenderableNodeSet) toIntermediate() []RenderableNode {
	intermediate := make([]RenderableNode, 0, n.Size())
	n.ForEach(func(node RenderableNode) {
		intermediate = append(intermediate, node)
	})
	return intermediate
}

func (n RenderableNodeSet) fromIntermediate(nodes []RenderableNode) RenderableNodeSet {
	return MakeRenderableNodeSet(nodes...)
}

// CodecEncodeSelf implements codec.Selfer
func (n *RenderableNodeSet) CodecEncodeSelf(encoder *codec.Encoder) {
	if n.psMap != nil {
		encoder.Encode(n.toIntermediate())
	} else {
		encoder.Encode(nil)
	}
}

// CodecDecodeSelf implements codec.Selfer
func (n *RenderableNodeSet) CodecDecodeSelf(decoder *codec.Decoder) {
	in := []RenderableNode{}
	if err := decoder.Decode(&in); err != nil {
		return
	}
	*n = RenderableNodeSet{}.fromIntermediate(in)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (RenderableNodeSet) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*RenderableNodeSet) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

// GobEncode implements gob.Marshaller
func (n RenderableNodeSet) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(n.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (n *RenderableNodeSet) GobDecode(input []byte) error {
	in := []RenderableNode{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*n = RenderableNodeSet{}.fromIntermediate(in)
	return nil
}
