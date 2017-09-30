package report

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/ugorji/go/codec"
)

type edgeMetadataEntry struct {
	key   string
	Value EdgeMetadata `json:"value"`
}

// EdgeMetadatas collect metadata about each edge in a topology. Keys are the
// remote node IDs, as in Adjacency.
type EdgeMetadatas struct {
	entries []edgeMetadataEntry
}

// MakeEdgeMetadatas returns EmptyEdgeMetadatas
func MakeEdgeMetadatas() EdgeMetadatas {
	return EdgeMetadatas{}
}

// locate the position where key should go, either at the end or in the middle
func (c EdgeMetadatas) locate(key string) int {
	return sort.Search(len(c.entries), func(i int) bool {
		return c.entries[i].key >= key
	})
}

// Add value to the counter 'key'
func (c EdgeMetadatas) Add(key string, value EdgeMetadata) EdgeMetadatas {
	i := c.locate(key)
	oldEntries := c.entries
	if i == len(c.entries) {
		c.entries = make([]edgeMetadataEntry, len(oldEntries)+1)
		copy(c.entries, oldEntries)
	} else if c.entries[i].key == key {
		value = value.Merge(c.entries[i].Value)
		c.entries = make([]edgeMetadataEntry, len(oldEntries))
		copy(c.entries, oldEntries)
	} else {
		c.entries = make([]edgeMetadataEntry, len(oldEntries)+1)
		copy(c.entries, oldEntries[:i])
		copy(c.entries[i+1:], oldEntries[i:])
	}
	c.entries[i] = edgeMetadataEntry{key: key, Value: value}
	return c
}

// Lookup the counter 'key'
func (c EdgeMetadatas) Lookup(key string) (EdgeMetadata, bool) {
	i := c.locate(key)
	if i < len(c.entries) && c.entries[i].key == key {
		return c.entries[i].Value, true
	}
	return EdgeMetadata{}, false
}

// Size is the number of elements
func (c EdgeMetadatas) Size() int {
	return len(c.entries)
}

// Merge produces a fresh Counters, container the keys from both inputs. When
// both inputs container the same key, the values are Merged.
func (m EdgeMetadatas) Merge(n EdgeMetadatas) EdgeMetadatas {
	switch {
	case m.entries == nil:
		return n
	case n.entries == nil:
		return m
	}
	out := make([]edgeMetadataEntry, 0, len(m.entries)+len(n.entries))

	i, j := 0, 0
	for i < len(m.entries) {
		switch {
		case j >= len(n.entries) || m.entries[i].key < n.entries[j].key:
			out = append(out, m.entries[i])
			i++
		case m.entries[i].key == n.entries[j].key:
			newValue := m.entries[i].Value.Merge(n.entries[j].Value)
			out = append(out, edgeMetadataEntry{key: m.entries[i].key, Value: newValue})
			i++
			j++
		default:
			out = append(out, n.entries[j])
			j++
		}
	}
	for ; j < len(n.entries); j++ {
		out = append(out, n.entries[j])
	}
	return EdgeMetadatas{out}
}

// Flatten flattens all the EdgeMetadatas in this set and returns the result.
// The original is not modified.
func (c EdgeMetadatas) Flatten() EdgeMetadata {
	result := EdgeMetadata{}
	c.ForEach(func(_ string, e EdgeMetadata) {
		result = result.Flatten(e)
	})
	return result
}

// ForEach executes f on each key value pair in the map
func (c EdgeMetadatas) ForEach(fn func(k string, v EdgeMetadata)) {
	for _, value := range c.entries {
		fn(value.key, value.Value)
	}
}

func (c EdgeMetadatas) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range c.entries {
		fmt.Fprintf(buf, "%s: %s,\n", val.key, val.Value)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other Counters
func (c EdgeMetadatas) DeepEqual(d EdgeMetadatas) bool {
	if c.Size() != d.Size() {
		return false
	}
	for i := range c.entries {
		if !reflect.DeepEqual(c.entries[i], d.entries[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer
// Note this uses undocumented, internal APIs, which could break
// in the future.  See https://github.com/weaveworks/scope/pull/1709
// for more information.
func (m *EdgeMetadatas) CodecEncodeSelf(encoder *codec.Encoder) {
	z, r := codec.GenHelperEncoder(encoder)
	if m.entries == nil {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(m.Size())
	for _, val := range m.entries {
		z.EncSendContainerState(containerMapKey)
		r.EncodeString(cUTF8, val.key)
		z.EncSendContainerState(containerMapValue)
		val.Value.CodecEncodeSelf(encoder)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer
// Uses undocumented, internal APIs as for CodecEncodeSelf.
func (m *EdgeMetadatas) CodecDecodeSelf(decoder *codec.Decoder) {
	m.entries = nil
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		m.entries = make([]edgeMetadataEntry, 0, length)
	}
	for i := 0; length < 0 || i < length; i++ {
		if length < 0 && r.CheckBreak() {
			break
		}
		z.DecSendContainerState(containerMapKey)
		var key string
		if !r.TryDecodeAsNil() {
			key = r.DecodeString()
		}
		i := m.locate(key)
		if i == len(m.entries) || m.entries[i].key != key {
			m.entries = append(m.entries, edgeMetadataEntry{})
			copy(m.entries[i+1:], m.entries[i:])
		}
		m.entries[i].key = key
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			m.entries[i].Value.CodecDecodeSelf(decoder)
		}
	}
	z.DecSendContainerState(containerMapEnd)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (EdgeMetadatas) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*EdgeMetadatas) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

// EdgeMetadata describes a superset of the metadata that probes can possibly
// collect about a directed edge between two nodes in any topology.
type EdgeMetadata struct {
	EgressPacketCount  *uint64 `json:"egress_packet_count,omitempty"`
	IngressPacketCount *uint64 `json:"ingress_packet_count,omitempty"`
	EgressByteCount    *uint64 `json:"egress_byte_count,omitempty"`  // Transport layer
	IngressByteCount   *uint64 `json:"ingress_byte_count,omitempty"` // Transport layer
	dummySelfer
}

// String returns a string representation of this EdgeMetadata
// Helps with our use of Spew and diff.
func (e EdgeMetadata) String() string {
	f := func(i *uint64) string {
		if i == nil {
			return "nil"
		}
		return strconv.FormatUint(*i, 10)
	}

	return fmt.Sprintf(`{
EgressPacketCount:  %v,
IngressPacketCount: %v,
EgressByteCount:    %v,
IngressByteCount:   %v,
}`,
		f(e.EgressPacketCount),
		f(e.IngressPacketCount),
		f(e.EgressByteCount),
		f(e.IngressByteCount))
}

// Copy returns a value copy of the EdgeMetadata.
func (e EdgeMetadata) Copy() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.EgressPacketCount),
		IngressPacketCount: cpu64ptr(e.IngressPacketCount),
		EgressByteCount:    cpu64ptr(e.EgressByteCount),
		IngressByteCount:   cpu64ptr(e.IngressByteCount),
	}
}

// Reversed returns a value copy of the EdgeMetadata, with the direction reversed.
func (e EdgeMetadata) Reversed() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.IngressPacketCount),
		IngressPacketCount: cpu64ptr(e.EgressPacketCount),
		EgressByteCount:    cpu64ptr(e.IngressByteCount),
		IngressByteCount:   cpu64ptr(e.EgressByteCount),
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
