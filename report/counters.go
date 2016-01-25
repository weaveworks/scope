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

// Counters is a string->int map.
type Counters struct {
	psMap ps.Map
}

// EmptyCounters is the set of empty counters.
var EmptyCounters = Counters{ps.NewMap()}

// MakeCounters returns EmptyCounters
func MakeCounters() Counters {
	return EmptyCounters
}

// Copy is a noop
func (c Counters) Copy() Counters {
	return c
}

// Add value to the counter 'key'
func (c Counters) Add(key string, value int) Counters {
	if c.psMap == nil {
		c = EmptyCounters
	}
	if existingValue, ok := c.psMap.Lookup(key); ok {
		value += existingValue.(int)
	}
	return Counters{
		c.psMap.Set(key, value),
	}
}

// Lookup the counter 'key'
func (c Counters) Lookup(key string) (int, bool) {
	if c.psMap != nil {
		existingValue, ok := c.psMap.Lookup(key)
		if ok {
			return existingValue.(int), true
		}
	}
	return 0, false
}

// Size returns the number of counters
func (c Counters) Size() int {
	if c.psMap == nil {
		return 0
	}
	return c.psMap.Size()
}

// Merge produces a fresh Counters, container the keys from both inputs. When
// both inputs container the same key, the latter value is used.
func (c Counters) Merge(other Counters) Counters {
	var (
		cSize     = c.Size()
		otherSize = other.Size()
		output    = c.psMap
		iter      = other.psMap
	)
	switch {
	case cSize == 0:
		return other
	case otherSize == 0:
		return c
	case cSize < otherSize:
		output, iter = iter, output
	}
	iter.ForEach(func(key string, otherVal interface{}) {
		if val, ok := output.Lookup(key); ok {
			output = output.Set(key, otherVal.(int)+val.(int))
		} else {
			output = output.Set(key, otherVal)
		}
	})
	return Counters{output}
}

func (c Counters) String() string {
	if c.psMap == nil {
		return "{}"
	}
	keys := []string{}
	for _, k := range c.psMap.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := c.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %d, ", key, val)
	}
	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

// DeepEqual tests equality with other Counters
func (c Counters) DeepEqual(i interface{}) bool {
	d, ok := i.(Counters)
	if !ok {
		return false
	}

	if (c.psMap == nil) != (d.psMap == nil) {
		return false
	} else if c.psMap == nil && d.psMap == nil {
		return true
	}

	if c.psMap.Size() != d.psMap.Size() {
		return false
	}

	equal := true
	c.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := d.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

func (c Counters) toIntermediate() map[string]int {
	intermediate := map[string]int{}
	c.psMap.ForEach(func(key string, val interface{}) {
		intermediate[key] = val.(int)
	})
	return intermediate
}

func (c Counters) fromIntermediate(in map[string]int) Counters {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return Counters{out}
}

// MarshalJSON implements json.Marshaller
func (c Counters) MarshalJSON() ([]byte, error) {
	if c.psMap != nil {
		return json.Marshal(c.toIntermediate())
	}
	return json.Marshal(nil)
}

// UnmarshalJSON implements json.Unmarshaler
func (c *Counters) UnmarshalJSON(input []byte) error {
	in := map[string]int{}
	if err := json.Unmarshal(input, &in); err != nil {
		return err
	}
	*c = Counters{}.fromIntermediate(in)
	return nil
}

// GobEncode implements gob.Marshaller
func (c Counters) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(c.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (c *Counters) GobDecode(input []byte) error {
	in := map[string]int{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*c = Counters{}.fromIntermediate(in)
	return nil
}
