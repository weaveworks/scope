package report

import (
	"reflect"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// Sets is a string->set-of-strings map.
// It is immutable.
type Sets struct {
	psMap ps.Map
}

// EmptySets is an empty Sets.  Starts with this.
var emptySets = Sets{ps.NewMap()}

// MakeSets returns EmptySets
func MakeSets() Sets {
	return emptySets
}

// Keys returns the keys for this set
func (s Sets) Keys() []string {
	if s.psMap == nil {
		return nil
	}
	return s.psMap.Keys()
}

// Add the given value to the Sets.
func (s Sets) Add(key string, value StringSet) Sets {
	if s.psMap == nil {
		s = emptySets
	}
	if existingValue, ok := s.psMap.Lookup(key); ok {
		var unchanged bool
		value, unchanged = existingValue.(StringSet).Merge(value)
		if unchanged {
			return s
		}
	}
	return Sets{
		psMap: s.psMap.Set(key, value),
	}
}

// AddString adds a single string under a key, creating a new StringSet if necessary.
func (s Sets) AddString(key string, str string) Sets {
	if s.psMap == nil {
		s = emptySets
	}
	value, found := s.Lookup(key)
	if found && value.Contains(str) {
		return s
	}
	value = value.Add(str)
	return Sets{
		psMap: s.psMap.Set(key, value),
	}
}

// Delete the given set from the Sets.
func (s Sets) Delete(key string) Sets {
	if s.psMap == nil {
		return emptySets
	}
	psMap := s.psMap.Delete(key)
	if psMap.IsNil() {
		return emptySets
	}
	return Sets{psMap: psMap}
}

// Lookup returns the sets stored under key.
func (s Sets) Lookup(key string) (StringSet, bool) {
	if s.psMap == nil {
		return MakeStringSet(), false
	}
	if value, ok := s.psMap.Lookup(key); ok {
		return value.(StringSet), true
	}
	return MakeStringSet(), false
}

// Size returns the number of elements
func (s Sets) Size() int {
	if s.psMap == nil {
		return 0
	}
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
			var unchanged bool
			set, unchanged = existingSet.(StringSet).Merge(set)
			if unchanged {
				return
			}
		}
		result = result.Set(key, set)
	})

	return Sets{result}
}

func (s Sets) String() string {
	return mapToString(s.psMap)
}

// DeepEqual tests equality with other Sets
func (s Sets) DeepEqual(t Sets) bool {
	return mapEqual(s.psMap, t.psMap, reflect.DeepEqual)
}

// CodecEncodeSelf implements codec.Selfer
func (s *Sets) CodecEncodeSelf(encoder *codec.Encoder) {
	mapWrite(s.psMap, encoder, func(encoder *codec.Encoder, val interface{}) {
		encoder.Encode(val.(StringSet))
	})
}

// CodecDecodeSelf implements codec.Selfer
func (s *Sets) CodecDecodeSelf(decoder *codec.Decoder) {
	out := mapRead(decoder, func(isNil bool) interface{} {
		var value StringSet
		if !isNil {
			decoder.Decode(&value)
		}
		return value
	})
	*s = Sets{out}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (Sets) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Sets) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
