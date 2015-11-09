package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
	return fmt.Sprintf("\"%s\" (%d)", e.Value, e.Timestamp)
}

// MakeLatestMap makes an empty LatestMap
func MakeLatestMap() LatestMap {
	return LatestMap{ps.NewMap()}
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

func (m LatestMap) toIntermediate() map[string]LatestEntry {
	intermediate := map[string]LatestEntry{}
	m.ForEach(func(key string, val interface{}) {
		intermediate[key] = val.(LatestEntry)
	})
	return intermediate
}

func fromIntermediate(in map[string]LatestEntry) LatestMap {
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
	*m = fromIntermediate(in)
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
	*m = fromIntermediate(in)
	return nil
}
