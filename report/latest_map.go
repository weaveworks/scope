package report

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"
	"time"

	"github.com/mndrix/ps"
	"github.com/ugorji/go/codec"
)

// LatestMap is a persitent map which support latest-win merges. We have to
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
	if m.Map == nil {
		return "", false
	}
	value, ok := m.Map.Lookup(key)
	if !ok {
		return "", false
	}
	return value.(LatestEntry).Value, true
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
	intermediate := map[string]LatestEntry{}
	if m.Map != nil {
		m.Map.ForEach(func(key string, val interface{}) {
			intermediate[key] = val.(LatestEntry)
		})
	}
	return intermediate
}

func (m LatestMap) fromIntermediate(in map[string]LatestEntry) LatestMap {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return LatestMap{out}
}

// CodecEncodeSelf implements codec.Selfer
func (m *LatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	if m.Map != nil {
		encoder.Encode(m.toIntermediate())
	} else {
		encoder.Encode(nil)
	}
}

// CodecDecodeSelf implements codec.Selfer
func (m *LatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	in := map[string]LatestEntry{}
	if err := decoder.Decode(&in); err != nil {
		return
	}
	*m = LatestMap{}.fromIntermediate(in)
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (LatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*LatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

// GobEncode implements gob.Marshaller
func (m LatestMap) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(m.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (m *LatestMap) GobDecode(input []byte) error {
	in := map[string]LatestEntry{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*m = LatestMap{}.fromIntermediate(in)
	return nil
}
