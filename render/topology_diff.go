package render

import (
	"reflect"
)

// Diff is returned by TopoDiff. It represents the changes between two
// RenderableNode maps.
type Diff struct {
	Add    []RenderableNode `json:"add"`
	Update []RenderableNode `json:"update"`
	Remove []string         `json:"remove"`
}

// TopoDiff gives you the diff to get from A to B.
func TopoDiff(a, b RenderableNodes) Diff {
	diff := Diff{}

	notSeen := map[string]struct{}{}
	a.ForEach(func(n RenderableNode) {
		notSeen[n.ID] = struct{}{}
	})

	b.ForEach(func(node RenderableNode) {
		if aNode, ok := a.Lookup(node.ID); !ok {
			diff.Add = append(diff.Add, node)
		} else if !reflect.DeepEqual(node, aNode) {
			diff.Update = append(diff.Update, node)
		}
		delete(notSeen, node.ID)
	})

	// leftover keys
	for k := range notSeen {
		diff.Remove = append(diff.Remove, k)
	}

	return diff
}
