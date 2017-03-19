package report

import (
	"fmt"
	"time"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// LatestEntryDecoder is an interface for decoding the LatestEntry instances.
type LatestEntryDecoder interface {
	Decode(decoder *codec.Decoder, entry *LatestEntry)
}

// LatestMap is a persistent map which support latest-win merges. We
// have to embed ps.Map as its interface. LatestMaps are immutable.
type LatestMap struct {
	ps.Map
	decoder LatestEntryDecoder
}

// LatestEntry represents a timestamped value inside the LatestMap.
type LatestEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     interface{} `json:"value"`
}

// String returns the LatestEntry's string representation.
func (e LatestEntry) String() string {
	return fmt.Sprintf("%v (%s)", e.Value, e.Timestamp.String())
}

// Equal returns true if the supplied LatestEntry is equal to this one.
func (e LatestEntry) Equal(e2 LatestEntry) bool {
	return e.Timestamp.Equal(e2.Timestamp) && e.Value == e2.Value
}

// MakeLatestMapWithDecoder makes an empty LatestMap holding custom values.
func MakeLatestMapWithDecoder(decoder LatestEntryDecoder) LatestMap {
	return LatestMap{ps.NewMap(), decoder}
}

// Copy is a noop, as LatestMaps are immutable.
func (m LatestMap) Copy() LatestMap {
	return m
}

// Size returns the number of elements.
func (m LatestMap) Size() int {
	if m.Map == nil {
		return 0
	}
	return m.Map.Size()
}

// Merge produces a fresh StringLatestMap containing the keys from
// both inputs. When both inputs contain the same key, the newer value
// is used.
func (m LatestMap) Merge(other LatestMap) LatestMap {
	if m.decoder != other.decoder {
		panic(fmt.Sprintf("Cannot merge maps with different entry value types, this has %#v, other has %#v", m.decoder, other.decoder))
	}
	output := mergeMaps(m.Map, other.Map, func(a, b interface{}) bool {
		return a.(LatestEntry).Timestamp.Before(b.(LatestEntry).Timestamp)
	})
	return LatestMap{output, m.decoder}
}

// Lookup the value for the given key.
func (m LatestMap) Lookup(key string) (interface{}, bool) {
	v, _, ok := m.LookupEntry(key)
	return v, ok
}

// LookupEntry returns the raw entry for the given key.
func (m LatestMap) LookupEntry(key string) (interface{}, time.Time, bool) {
	if m.Map == nil {
		return nil, time.Time{}, false
	}
	value, ok := m.Map.Lookup(key)
	if !ok {
		return nil, time.Time{}, false
	}
	e := value.(LatestEntry)
	return e.Value, e.Timestamp, true
}

// Set sets the value for the given key.
func (m LatestMap) Set(key string, timestamp time.Time, value interface{}) LatestMap {
	if m.Map == nil {
		m = MakeLatestMapWithDecoder(m.decoder)
	}
	return LatestMap{m.Map.Set(key, LatestEntry{timestamp, value}), m.decoder}
}

// Delete the value for the given key.
func (m LatestMap) Delete(key string) LatestMap {
	if m.Map == nil {
		m = MakeLatestMapWithDecoder(m.decoder)
	}
	return LatestMap{m.Map.Delete(key), m.decoder}
}

// ForEach executes fn on each key, timestamp, value triple in the map.
func (m LatestMap) ForEach(fn func(k string, ts time.Time, v interface{})) {
	if m.Map == nil {
		return
	}
	m.Map.ForEach(func(key string, value interface{}) {
		fn(key, value.(LatestEntry).Timestamp, value.(LatestEntry).Value)
	})
}

// String returns the LatestMap's string representation.
func (m LatestMap) String() string {
	return mapToString(m.Map)
}

// DeepEqual tests equality with other LatestMap.
func (m LatestMap) DeepEqual(n LatestMap) bool {
	if m.decoder != n.decoder {
		panic(fmt.Sprintf("Cannot check equality of maps with different entry value types, this has %#v, other has %#v", m.decoder, n.decoder))
	}
	return mapEqual(m.Map, n.Map, func(val, otherValue interface{}) bool {
		return val.(LatestEntry).Equal(otherValue.(LatestEntry))
	})
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

// CodecEncodeSelf implements codec.Selfer.
func (m *LatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	if m.Map != nil {
		encoder.Encode(m.toIntermediate())
	} else {
		encoder.Encode(nil)
	}
}

// CodecDecodeSelf implements codec.Selfer.
func (m *LatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	out := mapRead(decoder, func(isNil bool) interface{} {
		value := LatestEntry{}
		if !isNil {
			m.decoder.Decode(decoder, &value)
		}
		return value
	})
	*m = LatestMap{out, m.decoder}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead.
func (LatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead.
func (*LatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
