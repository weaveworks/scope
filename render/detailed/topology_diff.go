package detailed

import (
	"reflect"
)

// Diff is returned by TopoDiff. It represents the changes between two
// NodeSummary maps.
type Diff struct {
	Add    []NodeSummary `json:"add"`
	Update []NodeSummary `json:"update"`
	Remove []string      `json:"remove"`
	Reset  bool          `json:"reset,omitempty"`
}

// TopoDiff gives you the diff to get from A to B.
func TopoDiff(a, b NodeSummaries) Diff {
	diff := Diff{Reset: a == nil}

	notSeen := map[string]struct{}{}
	for k := range a {
		notSeen[k] = struct{}{}
	}

	for k, node := range b {
		if _, ok := a[k]; !ok {
			diff.Add = append(diff.Add, node)
		} else if !reflect.DeepEqual(node, a[k]) {
			diff.Update = append(diff.Update, node)
		}
		delete(notSeen, k)
	}

	// leftover keys
	for k := range notSeen {
		diff.Remove = append(diff.Remove, k)
	}

	return diff
}
