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
	// Handle nil/empty maps
	switch {
	case m.Size() == 0:
		return other
	case other.Size() == 0:
		return m
	}

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

// Intermediate format for serialization/de-serialization. We encode timestamps as
// nanoseconds since epoch to increase performance and use a type synonym for
// ps.Map so that we can provide a custom CodecEncodeSelf()/CodecDecodeSelf()
// just for the map without needing to worry about the timestamp field.
type intermediateMap struct {
	ps.Map
}

type intermediateConstMap struct {
	Map       intermediateMap `json:"map,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
}

func (m *intermediateMap) CodecDecodeSelf(decoder *codec.Decoder) {
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		m = nil
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

		var value string
		z.DecSendContainerState(containerMapValue)
		if !r.TryDecodeAsNil() {
			value = r.DecodeString()
		}

		out = out.UnsafeMutableSet(key, value)
	}
	m.Map = out
}

func (m *intermediateMap) CodecEncodeSelf(encoder *codec.Encoder) {
	z, r := codec.GenHelperEncoder(encoder)
	var length int
	if m.Map != nil {
		length = m.Map.Size()
	}
	if length == 0 {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(length)
	m.Map.ForEach(func(key string, iterVal interface{}) {
		z.EncSendContainerState(containerMapKey)
		r.EncodeString(cUTF83326, key)
		z.EncSendContainerState(containerMapValue)
		r.EncodeString(cUTF83326, iterVal.(string))
	})
	z.EncSendContainerState(containerMapEnd)
}

// CodecEncodeSelf implements codec.Selfer
func (m *ConstMap) CodecEncodeSelf(encoder *codec.Encoder) {
	// Only set the Timestamp field to non-zero if there is a map to
	// serialize, otherwise we would just be wasting space serializing a
	// non-zero timestamp.
	if m.Size() > 0 {
		var out intermediateConstMap
		out.Map.Map = m.Map
		out.Timestamp = m.Timestamp.UnixNano()
		encoder.Encode(out)
	} else {
		// Use an empty struct to avoid serializing the intermediate
		// "map" key
		encoder.Encode(struct{}{})
	}
}

// CodecDecodeSelf implements codec.Selfer
func (m *ConstMap) CodecDecodeSelf(decoder *codec.Decoder) {
	var in intermediateConstMap
	if err := decoder.Decode(&in); err != nil {
		return
	}
	if in.Map.Map == nil {
		*m = EmptyConstMap
	} else {
		m.Map = in.Map
		m.Timestamp = time.Unix(int64(0), in.Timestamp)
	}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (ConstMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*ConstMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
