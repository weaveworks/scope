// Generated file, do not edit.
// To regenerate, run ../extras/generate_latest_map ./latest_map_generated.go string NodeControlData

package report

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/ugorji/go/codec"
)

type stringLatestEntry struct {
	key       string
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
	dummySelfer
}

// String returns the StringLatestEntry's string representation.
func (e *stringLatestEntry) String() string {
	return fmt.Sprintf("%v (%s)", e.Value, e.Timestamp.String())
}

// Equal returns true if the supplied StringLatestEntry is equal to this one.
func (e *stringLatestEntry) Equal(e2 *stringLatestEntry) bool {
	return e.Timestamp.Equal(e2.Timestamp) && e.Value == e2.Value
}

// StringLatestMap holds latest string instances, as a slice sorted by key.
type StringLatestMap struct{ entries []stringLatestEntry }

// MakeStringLatestMap makes an empty StringLatestMap.
func MakeStringLatestMap() StringLatestMap {
	return StringLatestMap{}
}

// Size returns the number of elements.
func (m StringLatestMap) Size() int {
	return len(m.entries)
}

// Merge produces a fresh StringLatestMap containing the keys from both inputs.
// When both inputs contain the same key, the newer value is used.
func (m StringLatestMap) Merge(n StringLatestMap) StringLatestMap {
	switch {
	case m.entries == nil:
		return n
	case n.entries == nil:
		return m
	}
	out := make([]stringLatestEntry, 0, len(m.entries)+len(n.entries))

	i, j := 0, 0
	for i < len(m.entries) {
		switch {
		case j >= len(n.entries) || m.entries[i].key < n.entries[j].key:
			out = append(out, m.entries[i])
			i++
		case m.entries[i].key == n.entries[j].key:
			if m.entries[i].Timestamp.Before(n.entries[j].Timestamp) {
				out = append(out, n.entries[j])
			} else {
				out = append(out, m.entries[i])
			}
			i++
			j++
		default:
			out = append(out, n.entries[j])
			j++
		}
	}
	out = append(out, n.entries[j:]...)
	return StringLatestMap{out}
}

// Lookup the value for the given key.
func (m StringLatestMap) Lookup(key string) (string, bool) {
	v, _, ok := m.LookupEntry(key)
	if !ok {
		var zero string
		return zero, false
	}
	return v, true
}

// LookupEntry returns the raw entry for the given key.
func (m StringLatestMap) LookupEntry(key string) (string, time.Time, bool) {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	if i < len(m.entries) && m.entries[i].key == key {
		return m.entries[i].Value, m.entries[i].Timestamp, true
	}
	var zero string
	return zero, time.Time{}, false
}

// locate the position where key should go, and make room for it if not there already
func (m *StringLatestMap) locate(key string) int {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	if i == len(m.entries) || m.entries[i].key != key {
		m.entries = append(m.entries, stringLatestEntry{})
		copy(m.entries[i+1:], m.entries[i:])
	}
	return i
}

// Set the value for the given key.
func (m StringLatestMap) Set(key string, timestamp time.Time, value string) StringLatestMap {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	oldEntries := m.entries
	if i == len(m.entries) {
		m.entries = make([]stringLatestEntry, len(oldEntries)+1)
		copy(m.entries, oldEntries)
	} else if m.entries[i].key == key {
		m.entries = make([]stringLatestEntry, len(oldEntries))
		copy(m.entries, oldEntries)
	} else {
		m.entries = make([]stringLatestEntry, len(oldEntries)+1)
		copy(m.entries, oldEntries[:i])
		copy(m.entries[i+1:], oldEntries[i:])
	}
	m.entries[i] = stringLatestEntry{key: key, Timestamp: timestamp, Value: value}
	return m
}

// ForEach executes fn on each key value pair in the map.
func (m StringLatestMap) ForEach(fn func(k string, timestamp time.Time, v string)) {
	for _, value := range m.entries {
		fn(value.key, value.Timestamp, value.Value)
	}
}

// String returns the StringLatestMap's string representation.
func (m StringLatestMap) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m.entries {
		fmt.Fprintf(buf, "%s: %s,\n", val.key, val)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other StringLatestMap.
func (m StringLatestMap) DeepEqual(n StringLatestMap) bool {
	if m.Size() != n.Size() {
		return false
	}
	for i := range m.entries {
		if m.entries[i].key != n.entries[i].key || !m.entries[i].Equal(&n.entries[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer.
// Duplicates the output for a built-in map without generating an
// intermediate copy of the data structure, to save time.  Note this
// means we are using undocumented, internal APIs, which could break
// in the future.  See https://github.com/weaveworks/scope/pull/1709
// for more information.
func (m *StringLatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
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
		val.CodecEncodeSelf(encoder)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer.
// Decodes the input as for a built-in map, without creating an
// intermediate copy of the data structure to save time. Uses
// undocumented, internal APIs as for CodecEncodeSelf.
func (m *StringLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	m.entries = nil
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		m.entries = make([]stringLatestEntry, 0, length)
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
		m.entries[i].key = key
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			m.entries[i].CodecDecodeSelf(decoder)
		}
	}
	z.DecSendContainerState(containerMapEnd)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead.
func (StringLatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead.
func (*StringLatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

type nodeControlDataLatestEntry struct {
	key       string
	Timestamp time.Time       `json:"timestamp"`
	Value     NodeControlData `json:"value"`
	dummySelfer
}

// String returns the StringLatestEntry's string representation.
func (e *nodeControlDataLatestEntry) String() string {
	return fmt.Sprintf("%v (%s)", e.Value, e.Timestamp.String())
}

// Equal returns true if the supplied StringLatestEntry is equal to this one.
func (e *nodeControlDataLatestEntry) Equal(e2 *nodeControlDataLatestEntry) bool {
	return e.Timestamp.Equal(e2.Timestamp) && e.Value == e2.Value
}

// NodeControlDataLatestMap holds latest NodeControlData instances, as a slice sorted by key.
type NodeControlDataLatestMap struct{ entries []nodeControlDataLatestEntry }

// MakeNodeControlDataLatestMap makes an empty NodeControlDataLatestMap.
func MakeNodeControlDataLatestMap() NodeControlDataLatestMap {
	return NodeControlDataLatestMap{}
}

// Size returns the number of elements.
func (m NodeControlDataLatestMap) Size() int {
	return len(m.entries)
}

// Merge produces a fresh NodeControlDataLatestMap containing the keys from both inputs.
// When both inputs contain the same key, the newer value is used.
func (m NodeControlDataLatestMap) Merge(n NodeControlDataLatestMap) NodeControlDataLatestMap {
	switch {
	case m.entries == nil:
		return n
	case n.entries == nil:
		return m
	}
	out := make([]nodeControlDataLatestEntry, 0, len(m.entries)+len(n.entries))

	i, j := 0, 0
	for i < len(m.entries) {
		switch {
		case j >= len(n.entries) || m.entries[i].key < n.entries[j].key:
			out = append(out, m.entries[i])
			i++
		case m.entries[i].key == n.entries[j].key:
			if m.entries[i].Timestamp.Before(n.entries[j].Timestamp) {
				out = append(out, n.entries[j])
			} else {
				out = append(out, m.entries[i])
			}
			i++
			j++
		default:
			out = append(out, n.entries[j])
			j++
		}
	}
	out = append(out, n.entries[j:]...)
	return NodeControlDataLatestMap{out}
}

// Lookup the value for the given key.
func (m NodeControlDataLatestMap) Lookup(key string) (NodeControlData, bool) {
	v, _, ok := m.LookupEntry(key)
	if !ok {
		var zero NodeControlData
		return zero, false
	}
	return v, true
}

// LookupEntry returns the raw entry for the given key.
func (m NodeControlDataLatestMap) LookupEntry(key string) (NodeControlData, time.Time, bool) {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	if i < len(m.entries) && m.entries[i].key == key {
		return m.entries[i].Value, m.entries[i].Timestamp, true
	}
	var zero NodeControlData
	return zero, time.Time{}, false
}

// locate the position where key should go, and make room for it if not there already
func (m *NodeControlDataLatestMap) locate(key string) int {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	if i == len(m.entries) || m.entries[i].key != key {
		m.entries = append(m.entries, nodeControlDataLatestEntry{})
		copy(m.entries[i+1:], m.entries[i:])
	}
	return i
}

// Set the value for the given key.
func (m NodeControlDataLatestMap) Set(key string, timestamp time.Time, value NodeControlData) NodeControlDataLatestMap {
	i := sort.Search(len(m.entries), func(i int) bool {
		return m.entries[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	oldEntries := m.entries
	if i == len(m.entries) {
		m.entries = make([]nodeControlDataLatestEntry, len(oldEntries)+1)
		copy(m.entries, oldEntries)
	} else if m.entries[i].key == key {
		m.entries = make([]nodeControlDataLatestEntry, len(oldEntries))
		copy(m.entries, oldEntries)
	} else {
		m.entries = make([]nodeControlDataLatestEntry, len(oldEntries)+1)
		copy(m.entries, oldEntries[:i])
		copy(m.entries[i+1:], oldEntries[i:])
	}
	m.entries[i] = nodeControlDataLatestEntry{key: key, Timestamp: timestamp, Value: value}
	return m
}

// ForEach executes fn on each key value pair in the map.
func (m NodeControlDataLatestMap) ForEach(fn func(k string, timestamp time.Time, v NodeControlData)) {
	for _, value := range m.entries {
		fn(value.key, value.Timestamp, value.Value)
	}
}

// String returns the NodeControlDataLatestMap's string representation.
func (m NodeControlDataLatestMap) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m.entries {
		fmt.Fprintf(buf, "%s: %s,\n", val.key, val)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other NodeControlDataLatestMap.
func (m NodeControlDataLatestMap) DeepEqual(n NodeControlDataLatestMap) bool {
	if m.Size() != n.Size() {
		return false
	}
	for i := range m.entries {
		if m.entries[i].key != n.entries[i].key || !m.entries[i].Equal(&n.entries[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer.
// Duplicates the output for a built-in map without generating an
// intermediate copy of the data structure, to save time.  Note this
// means we are using undocumented, internal APIs, which could break
// in the future.  See https://github.com/weaveworks/scope/pull/1709
// for more information.
func (m *NodeControlDataLatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
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
		val.CodecEncodeSelf(encoder)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer.
// Decodes the input as for a built-in map, without creating an
// intermediate copy of the data structure to save time. Uses
// undocumented, internal APIs as for CodecEncodeSelf.
func (m *NodeControlDataLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	m.entries = nil
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		m.entries = make([]nodeControlDataLatestEntry, 0, length)
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
		m.entries[i].key = key
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			m.entries[i].CodecDecodeSelf(decoder)
		}
	}
	z.DecSendContainerState(containerMapEnd)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead.
func (NodeControlDataLatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead.
func (*NodeControlDataLatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
