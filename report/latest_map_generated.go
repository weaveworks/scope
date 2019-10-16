// Generated file, do not edit.
// To regenerate, run ../extras/generate_latest_map ./latest_map_generated.go string

package report

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ugorji/go/codec"
)

type stringLatestEntry struct {
	key   string
	value string
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

// Merge produces a StringLatestMap containing the keys from both inputs.
// m must be at least as new as n
// When both inputs contain the same key, the newer value is used.
// Tries to return one of its inputs, if that already holds the correct result.
func (m StringLatestMap) Merge(n StringLatestMap) StringLatestMap {
	switch {
	case len(m) == 0:
		return n
	case len(n) == 0:
		return m
	}

	i, j := 0, 0
loop:
	for i < len(m) {
		switch {
		case j >= len(n):
			return m
		case m[i].key == n[j].key:
			i++
			j++
		case m[i].key < n[j].key:
			i++
		default:
			break loop
		}
	}
	if i >= len(m) && j >= len(n) {
		return m
	}

	out := make([]stringLatestEntry, i, len(m))
	copy(out, m[:i])

	for i < len(m) {
		switch {
		case j >= len(n):
			out = append(out, m[i:]...)
			return out
		case m[i].key == n[j].key:
			out = append(out, m[i])
			i++
			j++
		case m[i].key < n[j].key:
			out = append(out, m[i])
			i++
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
	return m.LookupEntry(key)
}

// LookupEntry returns the raw entry for the given key.
// name exists for backwards-compatibility
func (m StringLatestMap) LookupEntry(key string) (string, bool) {
	i := sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
	if i < len(m) && m[i].key == key {
		return m[i].value, true
	}
	return "", false
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
		(*m)[i] = stringLatestEntry{}
	}
	return i
}

// Set the value for the given key.
func (m StringLatestMap) Set(key string, value string) StringLatestMap {
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
	m[i] = stringLatestEntry{key: key, value: value}
	return m
}

// ForEach executes fn on each key value pair in the map.
func (m StringLatestMap) ForEach(fn func(k string, v string)) {
	for _, value := range m {
		fn(value.key, value.value)
	}
}

// String returns the StringLatestMap's string representation.
func (m StringLatestMap) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m {
		fmt.Fprintf(buf, "%s: %s,\n", val.key, val.value)
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
		if m[i] != n[i] {
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
	z, r := codec.GenHelperEncoder(encoder)
	if m == nil {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(m.Size())
	for _, val := range m {
		z.EncSendContainerState(containerMapKey)
		r.EncodeString(cUTF8, val.key)
		z.EncSendContainerState(containerMapValue)
		r.EncodeString(cUTF8, val.value)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer.
// Decodes the input as for a built-in map, without creating an
// intermediate copy of the data structure to save time. Uses
// undocumented, internal APIs as for CodecEncodeSelf.
func (m *StringLatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	*m = nil
	z, r := codec.GenHelperDecoder(decoder)
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
		z.DecSendContainerState(containerMapKey)
		key := lookupCommonKey(r.DecodeStringAsBytes())
		i := m.locate(key)
		(*m)[i].key = key
		z.DecSendContainerState(containerMapValue)
		(*m)[i].value = r.DecodeString()
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
