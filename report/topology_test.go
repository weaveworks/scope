package report

import (
	"reflect"
	"sort"
	"testing"
)

func TestTopoDiff(t *testing.T) {
	nodea := RenderableNode{
		ID:         "nodea",
		LabelMajor: "Node A",
		LabelMinor: "'ts an a",
		Pseudo:     false,
		Adjacency: []string{
			"nodeb",
		},
	}
	nodeap := nodea
	nodeap.Adjacency = []string{
		"nodeb",
		"nodeq", // not the same anymore
	}
	nodeb := RenderableNode{
		ID:         "nodeb",
		LabelMajor: "Node B",
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...RenderableNode) RenderableNodes {
		r := RenderableNodes{}
		for _, n := range ns {
			r[n.ID] = n
		}
		return r
	}

	for _, c := range []struct {
		label      string
		have, want Diff
	}{
		{
			label: "basecase: empty -> something",
			have:  TopoDiff(nodes(), nodes(nodea, nodeb)),
			want: Diff{
				Add: []RenderableNode{nodea, nodeb},
			},
		},
		{
			label: "basecase: something -> empty",
			have:  TopoDiff(nodes(nodea, nodeb), nodes()),
			want: Diff{
				Remove: []string{"nodea", "nodeb"},
			},
		},
		{
			label: "add and remove",
			have:  TopoDiff(nodes(nodea), nodes(nodeb)),
			want: Diff{
				Add:    []RenderableNode{nodeb},
				Remove: []string{"nodea"},
			},
		},
		{
			label: "no change",
			have:  TopoDiff(nodes(nodea), nodes(nodea)),
			want:  Diff{},
		},
		{
			label: "change a single node",
			have:  TopoDiff(nodes(nodea), nodes(nodeap)),
			want: Diff{
				Update: []RenderableNode{nodeap},
			},
		},
	} {
		sort.Strings(c.have.Remove)
		sort.Sort(ByID(c.have.Add))
		sort.Sort(ByID(c.have.Update))
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - want:%s have:%s", c.label, c.want, c.have)
		}
	}
}
