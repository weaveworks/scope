package report

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// LatestMap is a persistent map which supports latest-win merges. We have to
// embed ps.Map as its an interface.  LatestMaps are immutable.
type LatestMap struct {
	ps.Map
}

// LatestEntry represents a timestamped value inside the LatestMap.
type LatestEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

func (e LatestEntry) String() string {
	return fmt.Sprintf("\"%s\" (%s)", e.Value, e.Timestamp.String())
}

// Equal returns true if the supplied LatestEntry is equal to this one.
func (e LatestEntry) Equal(e2 LatestEntry) bool {
	return e.Timestamp.Equal(e2.Timestamp) && e.Value == e2.Value
}

// EmptyLatestMap is an empty LatestMap.  Start with this.
var EmptyLatestMap = LatestMap{ps.NewMap()}

// MakeLatestMap makes an empty LatestMap
func MakeLatestMap() LatestMap {
	return EmptyLatestMap
}

// Copy is a noop, as LatestMaps are immutable.
func (m LatestMap) Copy() LatestMap {
	return m
}

// Size returns the number of elements
func (m LatestMap) Size() int {
	if m.Map == nil {
		return 0
	}
	return m.Map.Size()
}

// Merge produces a fresh LatestMap, container the kers from both inputs. When
// both inputs container the same key, the latter value is used.
func (m LatestMap) Merge(other LatestMap) LatestMap {
	var (
		mSize     = m.Size()
		otherSize = other.Size()
		output    = m.Map
		iter      = other.Map
	)
	switch {
	case mSize == 0:
		return other
	case otherSize == 0:
		return m
	case mSize < otherSize:
		output, iter = iter, output
	}

	iter.ForEach(func(key string, iterVal interface{}) {
		if existingVal, ok := output.Lookup(key); ok {
			if existingVal.(LatestEntry).Timestamp.Before(iterVal.(LatestEntry).Timestamp) {
				output = output.Set(key, iterVal)
			}
		} else {
			output = output.Set(key, iterVal)
		}
	})

	return LatestMap{output}
}

// Lookup the value for the given key.
func (m LatestMap) Lookup(key string) (string, bool) {
	v, _, ok := m.LookupEntry(key)
	return v, ok
}

// LookupEntry returns the raw entry for the given key.
func (m LatestMap) LookupEntry(key string) (string, time.Time, bool) {
	if m.Map == nil {
		return "", time.Time{}, false
	}
	value, ok := m.Map.Lookup(key)
	if !ok {
		return "", time.Time{}, false
	}
	e := value.(LatestEntry)
	return e.Value, e.Timestamp, true
}

// Set the value for the given key.
func (m LatestMap) Set(key string, timestamp time.Time, value string) LatestMap {
	if m.Map == nil {
		m = EmptyLatestMap
	}
	return LatestMap{m.Map.Set(key, LatestEntry{timestamp, value})}
}

// Delete the value for the given key.
func (m LatestMap) Delete(key string) LatestMap {
	if m.Map == nil {
		m = EmptyLatestMap
	}
	return LatestMap{m.Map.Delete(key)}
}

// ForEach executes f on each key value pair in the map
func (m LatestMap) ForEach(fn func(k, v string)) {
	if m.Map == nil {
		return
	}
	m.Map.ForEach(func(key string, value interface{}) {
		fn(key, value.(LatestEntry).Value)
	})
}

func (m LatestMap) String() string {
	keys := []string{}
	if m.Map == nil {
		m = EmptyLatestMap
	}
	for _, k := range m.Map.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := m.Map.Lookup(key)
		fmt.Fprintf(buf, "%s: %s,\n", key, val)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other LatestMap
func (m LatestMap) DeepEqual(n LatestMap) bool {
	if m.Size() != n.Size() {
		return false
	}
	if m.Size() == 0 {
		return true
	}

	equal := true
	m.Map.ForEach(func(k string, val interface{}) {
		if otherValue, ok := n.Map.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && val.(LatestEntry).Equal(otherValue.(LatestEntry))
		}
	})
	return equal
}

func (m LatestMap) toIntermediate() map[string]LatestEntry {
	intermediate := make(map[string]LatestEntry, m.Size())
	if m.Map != nil {
		m.Map.ForEach(func(key string, val interface{}) {
			intermediate[key] = val.(LatestEntry)
		})
	}
	return intermediate
}

// CodecEncodeSelf implements codec.Selfer
func (m *LatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	if m.Map != nil {
		encoder.Encode(m.toIntermediate())
	} else {
		encoder.Encode(nil)
	}
}

// constants from https://github.com/ugorji/go/blob/master/codec/helper.go#L207
const (
	containerMapKey   = 2
	containerMapValue = 3
	containerMapEnd   = 4
)

// CodecDecodeSelf implements codec.Selfer
// This implementation does not use the intermediate form as that was a
// performance issue; skipping it saved almost 10% CPU.  Note this means
// we are using undocumented, internal APIs, which could break in the future.
// See https://github.com/weaveworks/scope/pull/1709 for more information.
func (m *LatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		*m = LatestMap{}
		return
	}

	length := r.ReadMapStart()
	out := ps.NewMap()
	for i := 0; length < 0 || i < length; i++ {
		if length < 0 && r.CheckBreak() {
			break
		}

		var key string
		z.DecSendContainerState(containerMapKey)
		if !r.TryDecodeAsNil() {
			key = r.DecodeString()
		}

		var value LatestEntry
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			decoder.Decode(&value)
		}

		out = out.UnsafeMutableSet(key, value)
	}
	z.DecSendContainerState(containerMapEnd)
	*m = LatestMap{out}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (LatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*LatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
