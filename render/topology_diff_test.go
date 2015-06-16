package render

import (
	"reflect"
	"sort"
	"testing"

	"github.com/weaveworks/scope/report"
)

// ByID is a sort interface for a RenderableNode slice.
type ByID []report.RenderableNode

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].ID < r[j].ID }

func TestTopoDiff(t *testing.T) {
	nodea := report.RenderableNode{
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
	nodeb := report.RenderableNode{
		ID:         "nodeb",
		LabelMajor: "Node B",
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...report.RenderableNode) report.RenderableNodes {
		r := report.RenderableNodes{}
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
				Add: []report.RenderableNode{nodea, nodeb},
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
				Add:    []report.RenderableNode{nodeb},
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
				Update: []report.RenderableNode{nodeap},
			},
		},
	} {
		sort.Strings(c.have.Remove)
		sort.Sort(ByID(c.have.Add))
		sort.Sort(ByID(c.have.Update))
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - %s", c.label, diff(c.want, c.have))
		}
	}
}
