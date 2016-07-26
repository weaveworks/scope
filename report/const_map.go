package report

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// ConstMap is a persistent map which supports latest-win merges. Unlike
// LatestMap it doesn't support per-key timing granularity, it keeps a single
// timestamp to save space. We have to embed ps.Map as its an interface.
// ConstMaps are immutable.
type ConstMap struct {
	ps.Map
	Timestamp time.Time
}

// EmptyConstMap is an empty ConstMap.  Start with this.
var EmptyConstMap = ConstMap{ps.NewMap(), time.Time{}}

// MakeConstMap makes an empty ConstMap
func MakeConstMap() ConstMap {
	return EmptyConstMap
}

// Copy is a noop, as ConstMaps are immutable.
func (m ConstMap) Copy() ConstMap {
	return m
}

// Size returns the number of elements
func (m ConstMap) Size() int {
	if m.Map == nil {
		return 0
	}
	return m.Map.Size()
}

// Merge produces a fresh ConstMap, container the kers from both inputs. When
// both inputs container the same key, the latter value is used.
func (m ConstMap) Merge(other ConstMap) ConstMap {
	// Use the map with the latest timestamp as output
	// to minimize garbage
	latestTS, output, iter := m.Timestamp, m.Map, other.Map
	if other.Timestamp.After(latestTS) {
		latestTS, output, iter = other.Timestamp, other.Map, m.Map
	}

	iter.ForEach(func(key string, iterVal interface{}) {
		if _, ok := output.Lookup(key); !ok {
			output = output.Set(key, iterVal)
		}
	})

	return ConstMap{output, latestTS}
}

// Lookup the value for the given key.
func (m ConstMap) Lookup(key string) (string, bool) {
	if m.Map == nil {
		return "", false
	}
	value, ok := m.Map.Lookup(key)
	if !ok {
		return "", false
	}
	return value.(string), ok
}

// Set the value for the given key.
func (m ConstMap) Set(key string, timestamp time.Time, value string) ConstMap {
	if m.Map == nil {
		m = EmptyConstMap
	}
	return ConstMap{m.Map.Set(key, value), timestamp}
}

// Delete the value for the given key.
func (m ConstMap) Delete(key string) ConstMap {
	if m.Map == nil {
		m = EmptyConstMap
	}
	return ConstMap{m.Map.Delete(key), m.Timestamp}
}

// ForEach executes f on each key value pair in the map
func (m ConstMap) ForEach(fn func(k, v string)) {
	if m.Map == nil {
		return
	}
	m.Map.ForEach(func(key string, value interface{}) {
		fn(key, value.(string))
	})
}

func (m ConstMap) String() string {
	keys := []string{}
	if m.Map == nil {
		m = EmptyConstMap
	}
	for _, k := range m.Map.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("")
	fmt.Fprintf(buf, "{ %s, {\n", m.Timestamp)
	for _, key := range keys {
		val, _ := m.Map.Lookup(key)
		fmt.Fprintf(buf, "%s: %s,\n", key, val)
	}
	fmt.Fprintf(buf, "}\n}")
	return buf.String()
}

// DeepEqual tests equality with other ConstMap
func (m ConstMap) DeepEqual(n ConstMap) bool {
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
			equal = equal && val.(string) == otherValue.(string)
		}
	})
	return equal
}

type intermediateConstMap struct {
	Map       map[string]string `json:"map,omitempty"`
	Timestamp int64             `json:"timestamp,omitempty"`
}

func (m ConstMap) toIntermediate() intermediateConstMap {
	intermediate := intermediateConstMap{
		Map:       make(map[string]string, m.Map.Size()),
		Timestamp: m.Timestamp.UnixNano(),
	}
	if m.Map != nil {
		m.Map.ForEach(func(key string, val interface{}) {
			intermediate.Map[key] = val.(string)
		})
	}
	return intermediate
}

func (m *ConstMap) fromIntermediate(in intermediateConstMap) {
	m.Map = ps.NewMap()
	m.Timestamp = time.Unix(int64(0), in.Timestamp)
	for k, v := range in.Map {
		m.Map = m.Map.UnsafeMutableSet(k, v)
	}
}

// CodecEncodeSelf implements codec.Selfer
func (m *ConstMap) CodecEncodeSelf(encoder *codec.Encoder) {
	if m.Map != nil {
		encoder.Encode(m.toIntermediate())
	} else {
		encoder.Encode(nil)
	}
}

// CodecDecodeSelf implements codec.Selfer
func (m *ConstMap) CodecDecodeSelf(decoder *codec.Decoder) {
	var in intermediateConstMap
	if err := decoder.Decode(&in); err != nil {
		return
	}
	m.fromIntermediate(in)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (ConstMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*ConstMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
