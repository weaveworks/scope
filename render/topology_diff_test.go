package render_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/weaveworks/scope/render"
)

// ByID is a sort interface for a RenderableNode slice.
type ByID []render.RenderableNode

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].ID < r[j].ID }

func TestTopoDiff(t *testing.T) {
	nodea := render.RenderableNode{
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
	nodeb := render.RenderableNode{
		ID:         "nodeb",
		LabelMajor: "Node B",
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...render.RenderableNode) render.RenderableNodes {
		r := render.RenderableNodes{}
		for _, n := range ns {
			r[n.ID] = n
		}
		return r
	}

	for _, c := range []struct {
		label      string
		have, want render.Diff
	}{
		{
			label: "basecase: empty -> something",
			have:  render.TopoDiff(nodes(), nodes(nodea, nodeb)),
			want: render.Diff{
				Add: []render.RenderableNode{nodea, nodeb},
			},
		},
		{
			label: "basecase: something -> empty",
			have:  render.TopoDiff(nodes(nodea, nodeb), nodes()),
			want: render.Diff{
				Remove: []string{"nodea", "nodeb"},
			},
		},
		{
			label: "add and remove",
			have:  render.TopoDiff(nodes(nodea), nodes(nodeb)),
			want: render.Diff{
				Add:    []render.RenderableNode{nodeb},
				Remove: []string{"nodea"},
			},
		},
		{
			label: "no change",
			have:  render.TopoDiff(nodes(nodea), nodes(nodea)),
			want:  render.Diff{},
		},
		{
			label: "change a single node",
			have:  render.TopoDiff(nodes(nodea), nodes(nodeap)),
			want: render.Diff{
				Update: []render.RenderableNode{nodeap},
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
