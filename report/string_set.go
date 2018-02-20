package report

import (
	"sort"
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
func (s StringSet) Merge(other StringSet) StringSet {
	switch {
	case len(other) <= 0: // Optimise special case, to avoid allocating
		return s // (note unit test DeepEquals breaks if we don't do this)
	case len(s) <= 0:
		return other
	}
	result := make(StringSet, len(s)+len(other))
	for i, j, k := 0, 0, 0; ; k++ {
		switch {
		case i >= len(s):
			copy(result[k:], other[j:])
			return result[:k+len(other)-j]
		case j >= len(other):
			copy(result[k:], s[i:])
			return result[:k+len(s)-i]
		case s[i] < other[j]:
			result[k] = s[i]
			i++
		case s[i] > other[j]:
			result[k] = other[j]
			j++
		default: // equal
			result[k] = s[i]
			i++
			j++
		}
	}
}
