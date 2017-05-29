package report

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// Helper functions for ps.Map, without considering what is inside

// Return a new map containing all the elements of the two input maps
// and where the same key is in both, pick 'b' where prefer(a,b) is true
func mergeMaps(m, n ps.Map, prefer func(a, b interface{}) bool) ps.Map {
	switch {
	case m == nil:
		return n
	case n == nil:
		return m
	case m.Size() < n.Size():
		m, n = n, m
	}

	n.ForEach(func(key string, val interface{}) {
		if existingVal, found := m.Lookup(key); found {
			if prefer(existingVal, val) {
				m = m.Set(key, val)
			}
		} else {
			m = m.Set(key, val)
		}
	})

	return m
}

func mapEqual(m, n ps.Map, equalf func(a, b interface{}) bool) bool {
	var mSize, nSize int
	if m != nil {
		mSize = m.Size()
	}
	if n != nil {
		nSize = n.Size()
	}
	if mSize != nSize {
		return false
	}
	if mSize == 0 {
		return true
	}
	equal := true
	m.ForEach(func(k string, val interface{}) {
		if otherValue, ok := n.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && equalf(val, otherValue)
		}
	})
	return equal
}

// very similar to ps.Map.String() but with keys sorted
func mapToString(m ps.Map) string {
	buf := bytes.NewBufferString("{")
	for _, key := range mapKeys(m) {
		val, _ := m.Lookup(key)
		fmt.Fprintf(buf, "%s: %s,\n", key, val)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

func mapKeys(m ps.Map) []string {
	if m == nil {
		return nil
	}
	keys := m.Keys()
	sort.Strings(keys)
	return keys
}

// constants from https://github.com/ugorji/go/blob/master/codec/helper.go#L207
const (
	containerMapKey   = 2
	containerMapValue = 3
	containerMapEnd   = 4
	// from https://github.com/ugorji/go/blob/master/codec/helper.go#L152
	cUTF8 = 2
)

// This implementation does not use the intermediate form as that was a
// performance issue; skipping it saved almost 10% CPU.  Note this means
// we are using undocumented, internal APIs, which could break in the future.
// See https://github.com/weaveworks/scope/pull/1709 for more information.
func mapRead(decoder *codec.Decoder, decodeValue func(isNil bool) interface{}) ps.Map {
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		return ps.NewMap()
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

		z.DecSendContainerState(containerMapValue)
		value := decodeValue(r.TryDecodeAsNil())
		out = out.UnsafeMutableSet(key, value)
	}
	z.DecSendContainerState(containerMapEnd)
	return out
}

// Inverse of mapRead, done for performance. Same comments about
// undocumented internal APIs apply.
func mapWrite(m ps.Map, encoder *codec.Encoder, encodeValue func(*codec.Encoder, interface{})) {
	z, r := codec.GenHelperEncoder(encoder)
	if m == nil || m.IsNil() {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(m.Size())
	m.ForEach(func(key string, val interface{}) {
		z.EncSendContainerState(containerMapKey)
		r.EncodeString(cUTF8, key)
		z.EncSendContainerState(containerMapValue)
		encodeValue(encoder, val)
	})
	z.EncSendContainerState(containerMapEnd)
}
