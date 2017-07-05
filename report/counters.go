package report

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// Counters is a string->int map.
type Counters struct {
	psMap ps.Map
}

var emptyCounters = Counters{ps.NewMap()}

// MakeCounters returns EmptyCounters
func MakeCounters() Counters {
	return emptyCounters
}

// Add value to the counter 'key'
func (c Counters) Add(key string, value int) Counters {
	if c.psMap == nil {
		c = emptyCounters
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

// String serializes Counters into a string.
func (c Counters) String() string {
	buf := bytes.NewBufferString("{")
	prefix := ""
	for _, key := range mapKeys(c.psMap) {
		val, _ := c.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s%s: %d", prefix, key, val.(int))
		prefix = ", "
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other Counters
func (c Counters) DeepEqual(d Counters) bool {
	return mapEqual(c.psMap, d.psMap, reflect.DeepEqual)
}

func (c Counters) fromIntermediate(in map[string]int) Counters {
	out := ps.NewMap()
	for k, v := range in {
		out = out.Set(k, v)
	}
	return Counters{out}
}

// CodecEncodeSelf implements codec.Selfer
func (c *Counters) CodecEncodeSelf(encoder *codec.Encoder) {
	mapWrite(c.psMap, encoder, func(encoder *codec.Encoder, val interface{}) {
		i := val.(int)
		encoder.Encode(i)
	})
}

// CodecDecodeSelf implements codec.Selfer
func (c *Counters) CodecDecodeSelf(decoder *codec.Decoder) {
	out := mapRead(decoder, func(isNil bool) interface{} {
		var value int
		if !isNil {
			decoder.Decode(&value)
		}
		return value
	})
	*c = Counters{out}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (Counters) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Counters) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
