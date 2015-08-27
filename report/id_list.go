package report

import "sort"

// IDList is a list of string IDs, which are always sorted and unique.
type IDList []string

// MakeIDList makes a new IDList.
func MakeIDList(ids ...string) IDList {
	if len(ids) <= 0 {
		return IDList{}
	}
	sort.Strings(ids)
	for i := 1; i < len(ids); { // shuffle down any duplicates
		if ids[i-1] == ids[i] {
			ids = append(ids[:i-1], ids[i:]...)
			continue
		}
		i++
	}
	return IDList(ids)
}

// Add is the only correct way to add ids to an IDList.
func (a IDList) Add(ids ...string) IDList {
	for _, s := range ids {
		i := sort.Search(len(a), func(i int) bool { return a[i] >= s })
		if i < len(a) && a[i] == s {
			// The list already has the element.
			continue
		}
		// It a new element, insert it in order.
		a = append(a, "")
		copy(a[i+1:], a[i:])
		a[i] = s
	}
	return a
}

// Copy returns a copy of the IDList.
func (a IDList) Copy() IDList {
	result := make(IDList, len(a))
	copy(result, a)
	return result
}

// Merge all elements from a and b into a new list
func (a IDList) Merge(b IDList) IDList {
	if len(b) == 0 { // Optimise special case, to avoid allocating
		return a // (note unit test DeepEquals breaks if we don't do this)
	}
	d := make(IDList, len(a)+len(b))
	for i, j, k := 0, 0, 0; ; k++ {
		switch {
		case i >= len(a):
			copy(d[k:], b[j:])
			return d[:k+len(b)-j]
		case j >= len(b):
			copy(d[k:], a[i:])
			return d[:k+len(a)-i]
		case a[i] < b[j]:
			d[k] = a[i]
			i++
		case a[i] > b[j]:
			d[k] = b[j]
			j++
		default: // equal
			d[k] = a[i]
			i++
			j++
		}
	}
}

// Contains returns true if id is in the list.
func (a IDList) Contains(id string) bool {
	i := sort.Search(len(a), func(i int) bool { return a[i] >= id })
	return i < len(a) && a[i] == id
}
