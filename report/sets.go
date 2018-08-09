package report

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
)

// Sets holds StringSet instances, as a slice sorted by key.
type Sets []stringSetEntry

type stringSetEntry struct {
	key   string
	Value StringSet `json:"value"`
}

// MakeSets creates an empty Sets
func MakeSets() Sets {
	return nil
}

// locate the position where key should go, either at the end or in the middle
func (m Sets) locate(key string) int {
	return sort.Search(len(m), func(i int) bool {
		return m[i].key >= key
	})
}

// Add the given value to the Sets, merging the contents if the key is already there.
func (m Sets) Add(key string, value StringSet) Sets {
	i := m.locate(key)
	oldM := m
	if i == len(m) { // insert at end
		m = make([]stringSetEntry, len(oldM)+1)
		copy(m, oldM)
	} else if m[i].key == key { // merge with existing key
		var unchanged bool
		value, unchanged = m[i].Value.Merge(value)
		if unchanged {
			return m
		}
		m = make([]stringSetEntry, len(oldM))
		copy(m, oldM)
	} else { // insert in the middle
		m = make([]stringSetEntry, len(oldM)+1)
		copy(m, oldM[:i])
		copy(m[i+1:], oldM[i:])
	}
	m[i] = stringSetEntry{key: key, Value: value}
	return m
}

// AddString adds a single string under a key, creating a new StringSet if necessary.
func (m Sets) AddString(key string, str string) Sets {
	value, found := m.Lookup(key)
	if found && value.Contains(str) {
		return m
	}
	return m.Add(key, StringSet{str})
}

// Delete the given set from the Sets.
func (m Sets) Delete(key string) Sets {
	i := m.locate(key)
	if i == len(m) || m[i].key != key {
		return m // not found
	}
	result := make([]stringSetEntry, len(m)-1)
	if i > 0 {
		copy(result, m[:i-1])
	}
	copy(result[i:], m[i+1:])
	return result
}

// Lookup the value for the given key.
func (m Sets) Lookup(key string) (StringSet, bool) {
	i := m.locate(key)
	if i < len(m) && m[i].key == key {
		return m[i].Value, true
	}
	var zero StringSet
	return zero, false
}

// Size returns the number of elements
func (m Sets) Size() int {
	return len(m)
}

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (m Sets) Merge(n Sets) Sets {
	switch {
	case len(m) == 0:
		return n
	case len(n) == 0:
		return m
	}
	if len(n) > len(m) {
		m, n = n, m //swap so m is always at least as long as n
	}

	i, j := 0, 0
loop:
	for i < len(m) {
		switch {
		case j >= len(n):
			return m
		case m[i].key == n[j].key:
			if !m[i].Value.ContainsSet(n[j].Value) {
				break loop
			}
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

	out := make([]stringSetEntry, i, len(m))
	copy(out, m[:i])

	for i < len(m) {
		switch {
		case j >= len(n):
			out = append(out, m[i:]...)
			return out
		case m[i].key == n[j].key:
			newValue, _ := m[i].Value.Merge(n[j].Value)
			out = append(out, stringSetEntry{key: m[i].key, Value: newValue})
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

func (m Sets) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range m {
		fmt.Fprintf(buf, "%s: %v,\n", val.key, val.Value)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other Sets
func (m Sets) DeepEqual(n Sets) bool {
	if len(m) != len(n) {
		return false
	}
	for i := range m {
		if !reflect.DeepEqual(m[i], n[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer
// Note this uses undocumented, internal APIs, which could break
// in the future.  See https://github.com/weaveworks/scope/pull/1709
// for more information.
func (m Sets) CodecEncodeSelf(encoder *codec.Encoder) {
	z, r := codec.GenHelperEncoder(encoder)
	if m == nil {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(len(m))
	for _, val := range m {
		z.EncSendContainerState(containerMapKey)
		r.EncodeString(cUTF8, val.key)
		z.EncSendContainerState(containerMapValue)
		val.Value.CodecEncodeSelf(encoder)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer
// Uses undocumented, internal APIs as for CodecEncodeSelf.
func (m *Sets) CodecDecodeSelf(decoder *codec.Decoder) {
	*m = nil
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		*m = make([]stringSetEntry, 0, length)
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
		j := m.locate(key)
		if j == len(*m) || (*m)[j].key != key {
			*m = append(*m, stringSetEntry{})
			copy((*m)[j+1:], (*m)[j:])
			(*m)[j] = stringSetEntry{}
		}
		(*m)[j].key = key
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			(*m)[j].Value.CodecDecodeSelf(decoder)
		}
	}
	z.DecSendContainerState(containerMapEnd)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (Sets) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Sets) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
