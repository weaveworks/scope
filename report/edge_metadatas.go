package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mndrix/ps"
)

// EdgeMetadatas collect metadata about each edge in a topology. Keys are the
// remote node IDs, as in Adjacency.
type EdgeMetadatas struct {
	psMap ps.Map
}

// EmptyEdgeMetadatas is the set of empty EdgeMetadatas.
var EmptyEdgeMetadatas = EdgeMetadatas{ps.NewMap()}

// MakeEdgeMetadatas returns EmptyEdgeMetadatas
func MakeEdgeMetadatas() EdgeMetadatas {
	return EmptyEdgeMetadatas
}

// Copy is a noop
func (c EdgeMetadatas) Copy() EdgeMetadatas {
	return c
}

// Add value to the counter 'key'
func (c EdgeMetadatas) Add(key string, value EdgeMetadata) EdgeMetadatas {
	if existingValue, ok := c.psMap.Lookup(key); ok {
		value = value.Merge(existingValue.(EdgeMetadata))
	}
	return EdgeMetadatas{
		c.psMap.Set(key, value),
	}
}

// Lookup the counter 'key'
func (c EdgeMetadatas) Lookup(key string) (EdgeMetadata, bool) {
	existingValue, ok := c.psMap.Lookup(key)
	if ok {
		return existingValue.(EdgeMetadata), true
	}
	return EdgeMetadata{}, false
}

// Size is the number of elements
func (c EdgeMetadatas) Size() int {
	return c.psMap.Size()
}

// Merge produces a fresh Counters, container the keys from both inputs. When
// both inputs container the same key, the latter value is used.
func (c EdgeMetadatas) Merge(other EdgeMetadatas) EdgeMetadatas {
	var (
		cSize     = c.Size()
		otherSize = other.Size()
		output    = c.psMap
		iter      = other.psMap
	)
	switch {
	case cSize == 0:
		return other
	case otherSize == 0:
		return c
	case cSize < otherSize:
		output, iter = iter, output
	}
	iter.ForEach(func(key string, otherVal interface{}) {
		if val, ok := output.Lookup(key); ok {
			output = output.Set(key, otherVal.(EdgeMetadata).Merge(val.(EdgeMetadata)))
		} else {
			output = output.Set(key, otherVal)
		}
	})
	return EdgeMetadatas{output}
}

// Flatten flattens all the EdgeMetadatas in this set and returns the result.
// The original is not modified.
func (c EdgeMetadatas) Flatten() EdgeMetadata {
	result := EdgeMetadata{}
	c.psMap.ForEach(func(_ string, v interface{}) {
		result = result.Flatten(v.(EdgeMetadata))
	})
	return result
}

// ForEach executes f on each key value pair in the map
func (c EdgeMetadatas) ForEach(fn func(k string, v EdgeMetadata)) {
	c.psMap.ForEach(func(key string, value interface{}) {
		fn(key, value.(EdgeMetadata))
	})
}

func (c EdgeMetadatas) String() string {
	keys := []string{}
	for _, k := range c.psMap.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := c.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %v, ", key, val)
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// DeepEqual tests equality with other Counters
func (c EdgeMetadatas) DeepEqual(i interface{}) bool {
	d, ok := i.(EdgeMetadatas)
	if !ok {
		return false
	}

	if c.psMap.Size() != d.psMap.Size() {
		return false
	}

	equal := true
	c.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := d.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

func (c EdgeMetadatas) toIntermediate() map[string]EdgeMetadata {
	intermediate := map[string]EdgeMetadata{}
	c.psMap.ForEach(func(key string, val interface{}) {
		intermediate[key] = val.(EdgeMetadata)
	})
	return intermediate
}

func (c EdgeMetadatas) fromIntermediate(in map[string]EdgeMetadata) EdgeMetadatas {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return EdgeMetadatas{out}
}

// MarshalJSON implements json.Marshaller
func (c EdgeMetadatas) MarshalJSON() ([]byte, error) {
	if c.psMap != nil {
		return json.Marshal(c.toIntermediate())
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler
func (c *EdgeMetadatas) UnmarshalJSON(input []byte) error {
	in := map[string]EdgeMetadata{}
	if err := json.Unmarshal(input, &in); err != nil {
		return err
	}
	*c = EdgeMetadatas{}.fromIntermediate(in)
	return nil
}

// GobEncode implements gob.Marshaller
func (c EdgeMetadatas) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(c.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (c *EdgeMetadatas) GobDecode(input []byte) error {
	in := map[string]EdgeMetadata{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*c = EdgeMetadatas{}.fromIntermediate(in)
	return nil
}

// EdgeMetadata describes a superset of the metadata that probes can possibly
// collect about a directed edge between two nodes in any topology.
type EdgeMetadata struct {
	EgressPacketCount  *uint64 `json:"egress_packet_count,omitempty"`
	IngressPacketCount *uint64 `json:"ingress_packet_count,omitempty"`
	EgressByteCount    *uint64 `json:"egress_byte_count,omitempty"`  // Transport layer
	IngressByteCount   *uint64 `json:"ingress_byte_count,omitempty"` // Transport layer
	MaxConnCountTCP    *uint64 `json:"max_conn_count_tcp,omitempty"`
}

// Copy returns a value copy of the EdgeMetadata.
func (e EdgeMetadata) Copy() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.EgressPacketCount),
		IngressPacketCount: cpu64ptr(e.IngressPacketCount),
		EgressByteCount:    cpu64ptr(e.EgressByteCount),
		IngressByteCount:   cpu64ptr(e.IngressByteCount),
		MaxConnCountTCP:    cpu64ptr(e.MaxConnCountTCP),
	}
}

// Reversed returns a value copy of the EdgeMetadata, with the direction reversed.
func (e EdgeMetadata) Reversed() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.IngressPacketCount),
		IngressPacketCount: cpu64ptr(e.EgressPacketCount),
		EgressByteCount:    cpu64ptr(e.IngressByteCount),
		IngressByteCount:   cpu64ptr(e.EgressByteCount),
		MaxConnCountTCP:    cpu64ptr(e.MaxConnCountTCP),
	}
}

func cpu64ptr(u *uint64) *uint64 {
	if u == nil {
		return nil
	}
	value := *u   // oh man
	return &value // this sucks
}

// Merge merges another EdgeMetadata into the receiver and returns the result.
// The receiver is not modified. The two edge metadatas should represent the
// same edge on different times.
func (e EdgeMetadata) Merge(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, max)
	return cp
}

// Flatten sums two EdgeMetadatas and returns the result. The receiver is not
// modified. The two edge metadata windows should be the same duration; they
// should represent different edges at the same time.
func (e EdgeMetadata) Flatten(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	// Note that summing of two maximums doesn't always give us the true
	// maximum. But it's a best effort.
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, sum)
	return cp
}

func merge(dst, src *uint64, op func(uint64, uint64) uint64) *uint64 {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = new(uint64)
	}
	(*dst) = op(*dst, *src)
	return dst
}

func sum(dst, src uint64) uint64 {
	return dst + src
}

func max(dst, src uint64) uint64 {
	if dst > src {
		return dst
	}
	return src
}
