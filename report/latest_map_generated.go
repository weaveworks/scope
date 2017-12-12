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
type StringLatestMap []stringLatestEntry

// MakeStringLatestMap makes an empty StringLatestMap.
func MakeStringLatestMap() StringLatestMap {
	return StringLatestMap{}
}

// Size returns the number of elements.
func (m StringLatestMap) Size() int {
	return len(m)
}

// Merge produces a fresh StringLatestMap containing the keys from both inputs.
// When both inputs contain the same key, the newer value is used.
func (m StringLatestMap) Merge(n StringLatestMap) StringLatestMap {
	switch {
	case m == nil:
		return n
	case n == nil:
		return m
	}
	l := len(m)
	if len(n) > l {
		l = len(n)
	}
	out := make([]stringLatestEntry, 0, l)

	i, j := 0, 0
	for i < len(m) {
		switch {
		case j >= len(n) || m[i].key < n[j].key:
			out = append(out, m[i])
			i++
		case m[i].key == n[j].key:
			if m[i].Timestamp.Before(n[j].Timestamp) {
				out = append(out, n[j])
			} else {
				out = append(out, m[i])
			}
			i++
			j++
		default:
			out = append(out, n[j])
			j++
		}
	}
	out = append(out, n[j:]...)
	return out
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
	i := sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
	if i < len(m) && m[i].key == key {
		return m[i].Value, m[i].Timestamp, true
	}
	var zero string
	return zero, time.Time{}, false
}

// locate the position where key should go, and make room for it if not there already
func (m *StringLatestMap) locate(key string) int {
	i := sort.Search(len(*m), func(i int) bool {
		return (*m)[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	if i == len(*m) || (*m)[i].key != key {
		*m = append(*m, stringLatestEntry{})
		copy((*m)[i+1:], (*m)[i:])
	}
	return i
}

// Set the value for the given key.
func (m StringLatestMap) Set(key string, timestamp time.Time, value string) StringLatestMap {
	i := sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	oldEntries := m
	if i == len(m) {
		m = make([]stringLatestEntry, len(oldEntries)+1)
		copy(m, oldEntries)
	} else if m[i].key == key {
		m = make([]stringLatestEntry, len(oldEntries))
		copy(m, oldEntries)
	} else {
		m = make([]stringLatestEntry, len(oldEntries)+1)
		copy(m, oldEntries[:i])
		copy(m[i+1:], oldEntries[i:])
	}
	m[i] = stringLatestEntry{key: key, Timestamp: timestamp, Value: value}
	return m
}

// ForEach executes fn on each key value pair in the map.
func (m StringLatestMap) ForEach(fn func(k string, timestamp time.Time, v string)) {
	for _, value := range m {
		fn(value.key, value.Timestamp, value.Value)
	}
}

// String returns the StringLatestMap's string representation.
func (m StringLatestMap) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m {
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
	for i := range m {
		if m[i].key != n[i].key || !m[i].Equal(&n[i]) {
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
func (m StringLatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	_, r := codec.GenHelperEncoder(encoder)
	if m == nil {
		r.EncodeNil()
		return
	}
	r.WriteMapStart(m.Size())
	for _, val := range m {
		r.WriteMapElemKey()
		r.EncodeString(cUTF8, val.key)
		r.WriteMapElemValue()
		val.CodecEncodeSelf(encoder)
	}
	r.WriteMapEnd()
}

// CodecDecodeSelf implements codec.Selfer.
// Decodes the input as for a built-in map, without creating an
// intermediate copy of the data structure to save time. Uses
// undocumented, internal APIs as for CodecEncodeSelf.
func (m *StringLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	*m = nil
	_, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		*m = make([]stringLatestEntry, 0, length)
	}
	for i := 0; length < 0 || i < length; i++ {
		if length < 0 && r.CheckBreak() {
			break
		}
		r.ReadMapElemKey()
		var key string
		if !r.TryDecodeAsNil() {
			key = r.DecodeString()
		}
		i := m.locate(key)
		(*m)[i].key = key
		r.ReadMapElemValue()
		if !r.TryDecodeAsNil() {
			(*m)[i].CodecDecodeSelf(decoder)
		}
	}
	r.ReadMapEnd()
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
type NodeControlDataLatestMap []nodeControlDataLatestEntry

// MakeNodeControlDataLatestMap makes an empty NodeControlDataLatestMap.
func MakeNodeControlDataLatestMap() NodeControlDataLatestMap {
	return NodeControlDataLatestMap{}
}

// Size returns the number of elements.
func (m NodeControlDataLatestMap) Size() int {
	return len(m)
}

// Merge produces a fresh NodeControlDataLatestMap containing the keys from both inputs.
// When both inputs contain the same key, the newer value is used.
func (m NodeControlDataLatestMap) Merge(n NodeControlDataLatestMap) NodeControlDataLatestMap {
	switch {
	case m == nil:
		return n
	case n == nil:
		return m
	}
	l := len(m)
	if len(n) > l {
		l = len(n)
	}
	out := make([]nodeControlDataLatestEntry, 0, l)

	i, j := 0, 0
	for i < len(m) {
		switch {
		case j >= len(n) || m[i].key < n[j].key:
			out = append(out, m[i])
			i++
		case m[i].key == n[j].key:
			if m[i].Timestamp.Before(n[j].Timestamp) {
				out = append(out, n[j])
			} else {
				out = append(out, m[i])
			}
			i++
			j++
		default:
			out = append(out, n[j])
			j++
		}
	}
	out = append(out, n[j:]...)
	return out
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
	i := sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
	if i < len(m) && m[i].key == key {
		return m[i].Value, m[i].Timestamp, true
	}
	var zero NodeControlData
	return zero, time.Time{}, false
}

// locate the position where key should go, and make room for it if not there already
func (m *NodeControlDataLatestMap) locate(key string) int {
	i := sort.Search(len(*m), func(i int) bool {
		return (*m)[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	if i == len(*m) || (*m)[i].key != key {
		*m = append(*m, nodeControlDataLatestEntry{})
		copy((*m)[i+1:], (*m)[i:])
	}
	return i
}

// Set the value for the given key.
func (m NodeControlDataLatestMap) Set(key string, timestamp time.Time, value NodeControlData) NodeControlDataLatestMap {
	i := sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
	// i is now the position where key should go, either at the end or in the middle
	oldEntries := m
	if i == len(m) {
		m = make([]nodeControlDataLatestEntry, len(oldEntries)+1)
		copy(m, oldEntries)
	} else if m[i].key == key {
		m = make([]nodeControlDataLatestEntry, len(oldEntries))
		copy(m, oldEntries)
	} else {
		m = make([]nodeControlDataLatestEntry, len(oldEntries)+1)
		copy(m, oldEntries[:i])
		copy(m[i+1:], oldEntries[i:])
	}
	m[i] = nodeControlDataLatestEntry{key: key, Timestamp: timestamp, Value: value}
	return m
}

// ForEach executes fn on each key value pair in the map.
func (m NodeControlDataLatestMap) ForEach(fn func(k string, timestamp time.Time, v NodeControlData)) {
	for _, value := range m {
		fn(value.key, value.Timestamp, value.Value)
	}
}

// String returns the NodeControlDataLatestMap's string representation.
func (m NodeControlDataLatestMap) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m {
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
	for i := range m {
		if m[i].key != n[i].key || !m[i].Equal(&n[i]) {
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
func (m NodeControlDataLatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	_, r := codec.GenHelperEncoder(encoder)
	if m == nil {
		r.EncodeNil()
		return
	}
	r.WriteMapStart(m.Size())
	for _, val := range m {
		r.WriteMapElemKey()
		r.EncodeString(cUTF8, val.key)
		r.WriteMapElemValue()
		val.CodecEncodeSelf(encoder)
	}
	r.WriteMapEnd()
}

// CodecDecodeSelf implements codec.Selfer.
// Decodes the input as for a built-in map, without creating an
// intermediate copy of the data structure to save time. Uses
// undocumented, internal APIs as for CodecEncodeSelf.
func (m *NodeControlDataLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	*m = nil
	_, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		*m = make([]nodeControlDataLatestEntry, 0, length)
	}
	for i := 0; length < 0 || i < length; i++ {
		if length < 0 && r.CheckBreak() {
			break
		}
		r.ReadMapElemKey()
		var key string
		if !r.TryDecodeAsNil() {
			key = r.DecodeString()
		}
		i := m.locate(key)
		(*m)[i].key = key
		r.ReadMapElemValue()
		if !r.TryDecodeAsNil() {
			(*m)[i].CodecDecodeSelf(decoder)
		}
	}
	r.ReadMapEnd()
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead.
func (NodeControlDataLatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead.
func (*NodeControlDataLatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
