package report

import (
	"bytes"
	"fmt"
	"sort"
	"time"
	"unsafe"

	"github.com/ugorji/go/codec"
)

// NOTE: we would be able to get rid of the unsafe castings if
//       there was a way to compare strings and bytes without
//       conversion.

// (Maybe conversions will be efficient due what's explained in
// https://github.com/golang/go/issues/11777 but we need to call bytes.Compare()
// and I doubt escape analysis crosses function-call boundaries)

var (
	oneByteArr    = [1]byte{0}
	zeroByteSlice = oneByteArr[:0:0]
)

type unsafeString struct {
	Data uintptr
	Len  int
}

type unsafeSlice struct {
	Data uintptr
	Len  int
	Cap  int
}

// stringView returns a view of the []byte as a string.
// In unsafe mode, it doesn't incur allocation and copying caused by conversion.
// In regular safe mode, it is an allocation and copy.
func stringView(v []byte) string {
	if len(v) == 0 {
		return ""
	}

	bx := (*unsafeSlice)(unsafe.Pointer(&v))
	sx := unsafeString{bx.Data, bx.Len}
	return *(*string)(unsafe.Pointer(&sx))
}

// bytesView returns a view of the string as a []byte.
// In unsafe mode, it doesn't incur allocation and copying caused by conversion.
// In regular safe mode, it is an allocation and copy.
func bytesView(v string) []byte {
	if len(v) == 0 {
		return zeroByteSlice
	}

	sx := (*unsafeString)(unsafe.Pointer(&v))
	bx := unsafeSlice{sx.Data, sx.Len, sx.Len}
	return *(*[]byte)(unsafe.Pointer(&bx))
}

// TODO: probably too conservative, adjust later on
const slabSize = 128

// slabString implements a string, assuming most strings used will be smaller
// than slabSize. Replacing strings by slabStrings in arrays reduces allocations
// (and garbage collection pressure) while still supporting unbounded-size
// strings thanks to the slice extension
type slabString struct {
	Slab      [slabSize]byte
	SlabUsed  int
	Extension []byte
}

func (s slabString) String() string {
	if s.Extension == nil {
		return string(s.Slab[:s.SlabUsed])
	}
	return string(s.Bytes())
}

// TODO: consider calling the slabString methods below through pointer
//       if copies turn out to be too expensive

// Bytes obtains a contiguous copy of the slab string
func (s slabString) Bytes() []byte {
	if s.Extension == nil {
		result := make([]byte, s.SlabUsed)
		copy(result, s.Slab[:s.SlabUsed])
		return result
	}
	result := make([]byte, slabSize+len(s.Extension))
	copy(result, s.Slab[:])
	copy(result[slabSize:], s.Extension)
	return result
}

func (s *slabString) fromString(str string) {
	if len(str) <= slabSize {
		copy(s.Slab[:], str)
		s.SlabUsed = len(str)
		return
	}

	copy(s.Slab[:], str[:slabSize])
	s.SlabUsed = slabSize
	s.Extension = make([]byte, len(str)-slabSize)
	copy(s.Extension, str[slabSize:])
}

func (s *slabString) fromBytes(b []byte) {
	if len(b) <= slabSize {
		copy(s.Slab[:], b)
		s.SlabUsed = len(b)
		return
	}

	copy(s.Slab[:], b[:slabSize])
	s.SlabUsed = slabSize
	s.Extension = make([]byte, len(b)-slabSize)
	copy(s.Extension, b[slabSize:])
}

func (s slabString) Equal(other slabString) bool {
	// TODO: worth considering the extension case separately?
	return bytes.Equal(s.Slab[:s.SlabUsed], other.Slab[:other.SlabUsed]) &&
		bytes.Equal(s.Extension, other.Extension)
}

func (s slabString) compare(other slabString) int {
	slabCmp := bytes.Compare(s.Slab[:s.SlabUsed], other.Slab[:other.SlabUsed])

	if (s.Extension == nil && other.Extension == nil) || slabCmp != 0 {
		return slabCmp
	}

	return bytes.Compare(s.Extension, other.Extension)
}

// review this, thinking about prefixes of each other after the slabSize boundary
// (it's probably wrong)
func (s slabString) compareBytes(b []byte) int {
	if s.Extension == nil || len(b) < slabSize {
		return bytes.Compare(s.Slab[:s.SlabUsed], b)
	}

	if cmp := bytes.Compare(s.Slab[:], b[:slabSize]); cmp != 0 {
		return cmp
	}

	// slices are equal until slabSize, check the rest
	return bytes.Compare(s.Extension, b[slabSize:])
}

// CodecEncodeSelf implements codec.Selfer
func (s *slabString) CodecEncodeSelf(encoder *codec.Encoder) {
	_, r := codec.GenHelperEncoder(encoder)
	if s.Extension == nil {
		r.EncodeStringBytes(cUTF83326, s.Slab[:s.SlabUsed])
	} else {
		// Note: we could avoid the copy in Bytes() by encoding the slab
		// and extension in two invocations but the codec API doesn't
		// seem to support that
		r.EncodeStringBytes(cUTF83326, s.Bytes())
	}
}

// CodecDecodeSelf implements codec.Selfer
func (s *slabString) CodecDecodeSelf(decoder *codec.Decoder) {
	_, r := codec.GenHelperDecoder(decoder)
	b := r.DecodeBytes(nil, true, true)
	s.fromBytes(b)
}

// LatestMap is a persitent map which support latest-win merges.
type LatestMap []LatestEntry

// LatestEntry a key-value pair in the LatestMap.
type LatestEntry struct {
	Key   slabString
	Value LatestValue
}

// LatestValue represents a timestamped value inside the LatestMap.
type LatestValue struct {
	Timestamp time.Time  `json:"timestamp"`
	Value     slabString `json:"value"`
}

func (e LatestValue) String() string {
	return fmt.Sprintf("\"%s\" (%s)", e.Value.String(), e.Timestamp.String())
}

// Equal returns true if the supplied LatestValue is equal to this one.
func (e LatestValue) Equal(e2 LatestValue) bool {
	return e.Timestamp.Equal(e2.Timestamp) && e.Value.Equal(e2.Value)
}

// EmptyLatestMap is an empty LatestMap.  Start with this.
var EmptyLatestMap LatestMap = nil

// MakeLatestMap makes an empty LatestMap
// TODO: to be removed
func MakeLatestMap() LatestMap {
	return nil
}

// Copy copies the map
// TODO: is this needed externally? Can it be removed?
func (m LatestMap) Copy() LatestMap {
	result := make(LatestMap, len(m))
	// Do not deep-copy the extensions since they are
	// never exposed explicitly and cannot be mutated.
	copy(result, m)
	return result
}

// Size returns the number of elements
func (m LatestMap) Size() int {
	return len(m)
}

// Merge produces a fresh LatestMap, container the keys from both inputs. When
// both inputs contain the same key, the latter value is used.
func (m LatestMap) Merge(other LatestMap) LatestMap {

	switch {
	case len(m) == 0:
		return other
	case len(other) == 0:
		return m
	}

	// May be allocating too much space but append() allocates
	// space exponentially anyways.
	result := make(LatestMap, 0, len(m)+len(other))

	mI, otherI := 0, 0
	for {
		if otherI >= len(other) {
			result = append(result, m[mI:]...)
			break
		} else if mI >= len(m) {
			result = append(result, other[otherI:]...)
			break
		}

		// TODO: for better performance, instead of appending each
		//       individual element, we could detect and copy full subsegments.
		switch m[mI].Key.compare(other[otherI].Key) {
		case 0:
			if m[mI].Value.Timestamp.After(other[otherI].Value.Timestamp) {
				result = append(result, m[mI])
			} else {
				result = append(result, other[otherI])
			}
			mI++
			otherI++
		case -1:
			result = append(result, m[mI])
			mI++
		case 1:
			result = append(result, other[otherI])
			otherI++
		}
	}

	return result
}

// TODO: introduce slabString (or at least []byte) variants of Lookup and ForEach below
//       to avoid the expensive conversions back to string

// Lookup the value for the given key.
func (m LatestMap) Lookup(key string) (string, bool) {
	v, _, ok := m.LookupEntry(key)
	return v, ok
}

func (m LatestMap) lookupIndex(key []byte) int {
	return sort.Search(len(m), func(i int) bool {
		return m[i].Key.compareBytes(key) >= 0
	})
}

// LookupEntry returns the raw entry for the given key.
func (m LatestMap) LookupEntry(key string) (string, time.Time, bool) {
	if len(m) == 0 {
		return "", time.Time{}, false
	}

	b := bytesView(key)
	if i := m.lookupIndex(b); i < len(m) && m[i].Key.compareBytes(b) == 0 {
		// TODO: this is very inefficient due to the conversion to
		//       string
		return m[i].Value.Value.String(), m[i].Value.Timestamp, true
	}

	return "", time.Time{}, false
}

// Set the value for the given key.
func (m LatestMap) Set(key string, timestamp time.Time, value string) LatestMap {
	b := bytesView(key)
	i := m.lookupIndex(b)
	if i < len(m) && m[i].Key.compareBytes(b) == 0 {
		result := m.Copy()
		result[i].Key.fromString(key)
		result[i].Value.Value.fromString(value)
		result[i].Value.Timestamp = timestamp
		return result
	}

	// TODO: factor out the assignments (it's tempting to use a goto)
	result := make(LatestMap, len(m)+1)
	copy(result, m[:i])
	result[i].Key.fromString(key)
	result[i].Value.Value.fromString(value)
	result[i].Value.Timestamp = timestamp
	copy(result[i+1:], m[i:])
	return result
}

// Delete the value for the given key.
func (m LatestMap) Delete(key string) LatestMap {
	if len(m) == 0 {
		return m
	}

	b := bytesView(key)
	i := m.lookupIndex(b)
	if i > len(m) || m[i].Key.compareBytes(b) != 0 {
		return m
	}

	result := make(LatestMap, len(m)-1)
	copy(result, m[:i])
	copy(result[i:], m[i+1:])
	return result
}

// ForEach executes f on each key value pair in the map
func (m LatestMap) ForEach(fn func(k, v string)) {
	for i := 0; i < len(m); i++ {
		// TODO: this is very inefficient due to the conversion to
		//       strings, we could do an unsafe casting as long as the strings
		//       passed don't escape (which we can't guarantee for all cases)
		fn(m[i].Key.String(), m[i].Value.Value.String())
	}
}

// DeepEqual tests equality with other LatestMap
func (m LatestMap) DeepEqual(n LatestMap) bool {
	if len(m) != len(n) {
		return false
	}

	for i := 0; i < len(m); i++ {
		if !m[i].Key.Equal(n[i].Key) || !m[i].Value.Equal(n[i].Value) {
			return false
		}
	}

	return true
}

// constants from https://github.com/ugorji/go/blob/master/codec/helper.go#L207
const (
	// ----- content types ----
	cUTF83326 = 1
	// ----- containerStateValues ----
	containerMapKey   = 2
	containerMapValue = 3
	containerMapEnd   = 4
)

// CodecEncodeSelf implements codec.Selfer
func (m *LatestMap) CodecEncodeSelf(encoder *codec.Encoder) {
	z, r := codec.GenHelperEncoder(encoder)
	length := len(*m)
	if length == 0 {
		r.EncodeNil()
		return
	}
	r.EncodeMapStart(length)
	for i := 0; i < length; i++ {
		z.EncSendContainerState(containerMapKey)
		(*m)[i].Key.CodecEncodeSelf(encoder)
		z.EncSendContainerState(containerMapValue)
		encoder.Encode(&(*m)[i].Value)
	}
	z.EncSendContainerState(containerMapEnd)
}

// CodecDecodeSelf implements codec.Selfer
func (m *LatestMap) CodecDecodeSelf(decoder *codec.Decoder) {
	z, r := codec.GenHelperDecoder(decoder)
	if r.TryDecodeAsNil() {
		*m = nil
		return
	}

	// The length is present in MsgPack but not JSON
	var result LatestMap
	if length := r.ReadMapStart(); length >= 0 {
		result = make(LatestMap, length)
		for i := range result {
			z.DecSendContainerState(containerMapKey)
			result[i].Key.CodecDecodeSelf(decoder)
			z.DecSendContainerState(containerMapValue)
			decoder.Decode(&result[i].Value)
		}

	} else {
		var entry LatestEntry
		for !r.CheckBreak() {
			z.DecSendContainerState(containerMapKey)
			entry.Key.CodecDecodeSelf(decoder)
			z.DecSendContainerState(containerMapValue)
			decoder.Decode(&entry.Value)
			result = append(result, entry)
		}
	}
	z.DecSendContainerState(containerMapEnd)
	*m = result
	// TODO: sort map based on key values for backwards compatibility with older
	//       probes
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (LatestMap) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*LatestMap) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
