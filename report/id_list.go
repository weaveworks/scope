package report

// IDList is a list of string IDs, which are always sorted and unique.
type IDList StringSet

var emptyIDList = IDList(MakeStringSet())

// MakeIDList makes a new IDList.
func MakeIDList(ids ...string) IDList {
	if len(ids) == 0 {
		return emptyIDList
	}
	return IDList(MakeStringSet(ids...))
}

// Add is the only correct way to add ids to an IDList.
func (a IDList) Add(ids ...string) IDList {
	if len(ids) == 0 {
		return a
	}
	return IDList(StringSet(a).Add(ids...))
}

// Merge all elements from a and b into a new list
func (a IDList) Merge(b IDList) IDList {
	merged, _ := StringSet(a).Merge(StringSet(b))
	return IDList(merged)
}

// Contains returns true if id is in the list.
func (a IDList) Contains(id string) bool {
	return StringSet(a).Contains(id)
}

// Intersection returns the intersection of a and b
func (a IDList) Intersection(b IDList) IDList {
	return IDList(StringSet(a).Intersection(StringSet(b)))
}

// Minus returns the set with id removed
func (a IDList) Minus(id string) IDList {
	return IDList(StringSet(a).Minus(id))
}
