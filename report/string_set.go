package report

import (
	"sort"

	"github.com/ugorji/go/codec"
)

// StringSet is a sorted set of unique strings. Clients must use the Add
// method to add strings.
type StringSet []string

var emptyStringSet StringSet

// MakeStringSet makes a new StringSet with the given strings.
func MakeStringSet(strs ...string) StringSet {
	if len(strs) <= 0 {
		return nil
	}
	result := make([]string, len(strs))
	copy(result, strs)
	sort.Strings(result)
	for i := 1; i < len(result); { // shuffle down any duplicates
		if result[i-1] == result[i] {
			result = append(result[:i-1], result[i:]...)
			continue
		}
		i++
	}
	return StringSet(result)
}

// Contains returns true if the string set includes the given string
func (s StringSet) Contains(str string) bool {
	i := sort.Search(len(s), func(i int) bool { return s[i] >= str })
	return i < len(s) && s[i] == str
}

// ContainsSet returns true if the string set includes all strings in the other set
func (s StringSet) ContainsSet(b StringSet) bool {
	i, j := 0, 0
	for i < len(s) && j < len(b) {
		if s[i] == b[j] {
			i++
			j++
		} else if s[i] < b[j] {
			i++
		} else {
			return false
		}
	}
	return j == len(b)
}

// Intersection returns the intersections of a and b
func (s StringSet) Intersection(b StringSet) StringSet {
	result, i, j := emptyStringSet, 0, 0
	for i < len(s) && j < len(b) {
		if s[i] == b[j] {
			result = result.Add(s[i])
		}
		if s[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return result
}

// Equal returns true if a and b have the same contents
func (s StringSet) Equal(b StringSet) bool {
	if len(s) != len(b) {
		return false
	}
	for i := range s {
		if s[i] != b[i] {
			return false
		}
	}
	return true
}

// Add adds the strings to the StringSet. Add is the only valid way to grow a
// StringSet. Add returns the StringSet to enable chaining.
func (s StringSet) Add(strs ...string) StringSet {
	for _, str := range strs {
		i := sort.Search(len(s), func(i int) bool { return s[i] >= str })
		if i < len(s) && s[i] == str {
			// The list already has the element.
			continue
		}
		// It a new element, insert it in order.
		s = append(s, "")
		copy(s[i+1:], s[i:])
		s[i] = str
	}
	return s
}

// Merge combines the two StringSets and returns a new result.
// Second return value is true if the return value is s
func (s StringSet) Merge(other StringSet) (StringSet, bool) {
	switch {
	case len(other) <= 0: // Optimise special case, to avoid allocating
		return s, true // (note unit test DeepEquals breaks if we don't do this)
	case len(s) <= 0:
		return other, false
	}

	i, j := 0, 0
loop:
	for i < len(s) {
		switch {
		case j >= len(other):
			return s, true
		case s[i] == other[j]:
			i++
			j++
		case s[i] < other[j]:
			i++
		default:
			break loop
		}
	}
	if i >= len(s) && j >= len(other) {
		return s, true
	}

	result := make(StringSet, i, len(s)+len(other))
	copy(result, s[:i])

	for i < len(s) {
		switch {
		case j >= len(other):
			result = append(result, s[i:]...)
			return result, false
		case s[i] == other[j]:
			result = append(result, s[i])
			i++
			j++
		case s[i] < other[j]:
			result = append(result, s[i])
			i++
		default:
			result = append(result, other[j])
			j++
		}
	}
	result = append(result, other[j:]...)
	return result, false
}

// CodecEncodeSelf implements codec.Selfer
func (s StringSet) CodecEncodeSelf(encoder *codec.Encoder) {
	z, r := codec.GenHelperEncoder(encoder)
	if s == nil {
		r.EncodeNil()
		return
	}
	r.EncodeArrayStart(len(s))
	for _, yyv1 := range s {
		z.EncSendContainerState(containerArrayElem)
		r.EncodeString(cUTF8, yyv1)
	}
	z.EncSendContainerState(containerArrayEnd)
}

// CodecDecodeSelf implements codec.Selfer
func (s *StringSet) CodecDecodeSelf(decoder *codec.Decoder) {
	z, r := codec.GenHelperDecoder(decoder)
	yyh1, length := z.DecSliceHelperStart()
	if length == 0 {
		*s = []string{}
		return
	} else if length < 0 {
		*s = make([]string, 0, 8)
	} else {
		*s = make([]string, 0, length)
	}
	for i := 0; length < 0 || i < length; i++ {
		if length < 0 && r.CheckBreak() {
			break
		}
		yyh1.ElemContainerState(i)
		str := r.DecodeString()
		*s = append(*s, str)
	}
	yyh1.End()
}
