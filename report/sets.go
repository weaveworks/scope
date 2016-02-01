package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mndrix/ps"
)

// Sets is a string->set-of-strings map.
// It is immutable.
type Sets struct {
	psMap ps.Map
}

// EmptySets is an empty Sets.  Starts with this.
var EmptySets = Sets{ps.NewMap()}

// MakeSets returns EmptySets
func MakeSets() Sets {
	return EmptySets
}

// Keys returns the keys for this set
func (s Sets) Keys() []string {
	return s.psMap.Keys()
}

// Add the given value to the Sets.
func (s Sets) Add(key string, value StringSet) Sets {
	if existingValue, ok := s.psMap.Lookup(key); ok {
		value = value.Merge(existingValue.(StringSet))
	}
	return Sets{
		psMap: s.psMap.Set(key, value),
	}
}

// Lookup returns the sets stored under key.
func (s Sets) Lookup(key string) (StringSet, bool) {
	if value, ok := s.psMap.Lookup(key); ok {
		return value.(StringSet), true
	}
	return EmptyStringSet, false
}

// Size returns the number of elements
func (s Sets) Size() int {
	return s.psMap.Size()
}

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (s Sets) Merge(other Sets) Sets {
	var (
		sSize     = s.Size()
		otherSize = other.Size()
		result    = s.psMap
		iter      = other.psMap
	)
	switch {
	case sSize == 0:
		return other
	case otherSize == 0:
		return s
	case sSize < otherSize:
		result, iter = iter, result
	}

	iter.ForEach(func(key string, value interface{}) {
		set := value.(StringSet)
		if existingSet, ok := result.Lookup(key); ok {
			set = set.Merge(existingSet.(StringSet))
		}
		result = result.Set(key, set)
	})

	return Sets{result}
}

// Copy is a noop
func (s Sets) Copy() Sets {
	return s
}

func (s Sets) String() string {
	keys := []string{}
	for _, k := range s.psMap.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := s.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %v, ", key, val)
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// DeepEqual tests equality with other Sets
func (s Sets) DeepEqual(t Sets) bool {
	if s.psMap.Size() != t.psMap.Size() {
		return false
	}

	equal := true
	s.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := t.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

func (s Sets) toIntermediate() map[string]StringSet {
	intermediate := map[string]StringSet{}
	s.psMap.ForEach(func(key string, val interface{}) {
		intermediate[key] = val.(StringSet)
	})
	return intermediate
}

func (s Sets) fromIntermediate(in map[string]StringSet) Sets {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return Sets{out}
}

// MarshalJSON implements json.Marshaller
func (s Sets) MarshalJSON() ([]byte, error) {
	if s.psMap != nil {
		return json.Marshal(s.toIntermediate())
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Sets) UnmarshalJSON(input []byte) error {
	in := map[string]StringSet{}
	if err := json.Unmarshal(input, &in); err != nil {
		return err
	}
	*s = Sets{}.fromIntermediate(in)
	return nil
}

// GobEncode implements gob.Marshaller
func (s Sets) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(s.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (s *Sets) GobDecode(input []byte) error {
	in := map[string]StringSet{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*s = Sets{}.fromIntermediate(in)
	return nil
}
