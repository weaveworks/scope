package report

import (
	"sort"
)

// IDList is a list of string IDs, which are always sorted and
// without duplicates.
type IDList []string

// NewIDList makes a new IDList.
func NewIDList(ids ...string) IDList {
	sort.Strings(ids)
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
		a = append(a[:i], append(IDList{s}, a[i:]...)...)
	}
	return a
}
