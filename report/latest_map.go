package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mndrix/ps"
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
	return fmt.Sprintf("\"%s\" (%d)", e.Value, e.Timestamp.String())
}

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

// Merge produces a fresh LatestMap, container the kers from both inputs. When
// both inputs container the same key, the latter value is used.
func (m LatestMap) Merge(newer LatestMap) LatestMap {
	// expect people to do old.Merge(new), optimise for that.
	// ie if you do {k: v}.Merge({k: v'}), we end up just returning
	// newer, unmodified.
	output := newer.Map

	m.Map.ForEach(func(key string, olderVal interface{}) {
		if newerVal, ok := newer.Map.Lookup(key); ok {
			if newerVal.(LatestEntry).Timestamp.Before(olderVal.(LatestEntry).Timestamp) {
				output = output.Set(key, olderVal)
			}
		} else {
			output = output.Set(key, olderVal)
		}
	})

	return LatestMap{output}
}

// Lookup the value for the given key.
func (m LatestMap) Lookup(key string) (string, bool) {
	value, ok := m.Map.Lookup(key)
	if !ok {
		return "", false
	}
	return value.(LatestEntry).Value, true
}

// Set the value for the given key.
func (m LatestMap) Set(key string, timestamp time.Time, value string) LatestMap {
	return LatestMap{m.Map.Set(key, LatestEntry{timestamp, value})}
}

// Delete the value for the given key.
func (m LatestMap) Delete(key string) LatestMap {
	return LatestMap{m.Map.Delete(key)}
}

// ForEach executes f on each key value pair in the map
func (m LatestMap) ForEach(fn func(k, v string)) {
	m.Map.ForEach(func(key string, value interface{}) {
		fn(key, value.(LatestEntry).Value)
	})
}

func (m LatestMap) String() string {
	keys := []string{}
	for _, k := range m.Map.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := m.Map.Lookup(key)
		fmt.Fprintf(buf, "%s: %s, ", key, val)
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// DeepEqual tests equality with other LatestMap
func (m LatestMap) DeepEqual(i interface{}) bool {
	n, ok := i.(LatestMap)
	if !ok {
		return false
	}

	if m.Map.Size() != n.Map.Size() {
		return false
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
	m.Map.ForEach(func(key string, val interface{}) {
		intermediate[key] = val.(LatestEntry)
	})
	return intermediate
}

func (m LatestMap) fromIntermediate(in map[string]LatestEntry) LatestMap {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return LatestMap{out}
}

// MarshalJSON implements json.Marshaller
func (m LatestMap) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	var err error
	if m.Map != nil {
		err = json.NewEncoder(&buf).Encode(m.toIntermediate())
	} else {
		err = json.NewEncoder(&buf).Encode(nil)
	}
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (m *LatestMap) UnmarshalJSON(input []byte) error {
	in := map[string]LatestEntry{}
	if err := json.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*m = LatestMap{}.fromIntermediate(in)
	return nil
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
