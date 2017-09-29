package report

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
)

type stringSetEntry struct {
	key   string
	Value StringSet `json:"value"`
}

// Sets is a string->set-of-strings map.
type Sets struct {
	entries []stringSetEntry
}

// MakeSets returns EmptySets
func MakeSets() Sets {
	return Sets{}
}

// Keys returns the keys for this set
func (s Sets) Keys() []string {
	keys := make([]string, s.Size())
	for i, v := range s.entries {
		keys[i] = v.key
	}
	return keys
}

// locate the position where key should go, either at the end or in the middle
func (s Sets) locate(key string) int {
	return sort.Search(len(s.entries), func(i int) bool {
		return s.entries[i].key >= key
	})
}

// Add the given value to the Sets, merging the contents if the key is already there.
func (s Sets) Add(key string, value StringSet) Sets {
	i := s.locate(key)
	oldEntries := s.entries
	if i == len(s.entries) {
		s.entries = make([]stringSetEntry, len(oldEntries)+1)
		copy(s.entries, oldEntries)
	} else if s.entries[i].key == key {
		value = value.Merge(s.entries[i].Value)
		s.entries = make([]stringSetEntry, len(oldEntries))
		copy(s.entries, oldEntries)
	} else {
		s.entries = make([]stringSetEntry, len(oldEntries)+1)
		copy(s.entries, oldEntries[:i])
		copy(s.entries[i+1:], oldEntries[i:])
	}
	s.entries[i] = stringSetEntry{key: key, Value: value}
	return s
}

// Delete the given set from the Sets.
func (s Sets) Delete(key string) Sets {
	i := s.locate(key)
	if i == len(s.entries) || s.entries[i].key != key {
		return s // not found
	}
	result := make([]stringSetEntry, len(s.entries)-1)
	copy(result, s.entries[:i-1])
	copy(result[i:], s.entries[i+1:])
	return Sets{entries: result}
}

// Lookup returns the sets stored under key.
func (s Sets) Lookup(key string) (StringSet, bool) {
	i := s.locate(key)
	if i < len(s.entries) && s.entries[i].key == key {
		return s.entries[i].Value, true
	}
	return MakeStringSet(), false
}

// Size returns the number of elements
func (s Sets) Size() int {
	return len(s.entries)
}

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (m Sets) Merge(n Sets) Sets {
	switch {
	case m.entries == nil:
		return n
	case n.entries == nil:
		return m
	}
	out := make([]stringSetEntry, 0, len(m.entries)+len(n.entries))

	i, j := 0, 0
	for i < len(m.entries) {
		switch {
		case j >= len(n.entries) || m.entries[i].key < n.entries[j].key:
			out = append(out, m.entries[i])
			i++
		case m.entries[i].key == n.entries[j].key:
			newValue := m.entries[i].Value.Merge(n.entries[j].Value)
			out = append(out, stringSetEntry{key: m.entries[i].key, Value: newValue})
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
	return Sets{out}
}

func (s Sets) String() string {
	buf := bytes.NewBufferString("{")
	for _, val := range s.entries {
		fmt.Fprintf(buf, "%s: %s,\n", val.key, val.Value)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other Sets
func (s Sets) DeepEqual(t Sets) bool {
	if s.Size() != t.Size() {
		return false
	}
	for i := range s.entries {
		if !reflect.DeepEqual(s.entries[i], t.entries[i]) {
			return false
		}
	}
	return true
}

// CodecEncodeSelf implements codec.Selfer
// Note this uses undocumented, internal APIs, which could break
// in the future.  See https://github.com/weaveworks/scope/pull/1709
// for more information.
func (m *Sets) CodecEncodeSelf(encoder *codec.Encoder) {
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
func (m *Sets) CodecDecodeSelf(decoder *codec.Decoder) {
	m.entries = nil
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return
	}

	length := r.ReadMapStart()
	if length > 0 {
		m.entries = make([]stringSetEntry, 0, length)
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
			m.entries = append(m.entries, stringSetEntry{})
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
func (Sets) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Sets) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
