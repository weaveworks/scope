package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestFilterRender(t *testing.T) {
	renderer := render.FilterUnconnected(
		mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode().WithID("foo").WithAdjacent("bar"),
			"bar": report.MakeNode().WithID("bar").WithAdjacent("foo"),
			"baz": report.MakeNode().WithID("baz"),
		}})

	have := report.MakeIDList()
	for id := range renderer.Render(report.MakeReport()) {
		have = have.Add(id)
	}
	want := report.MakeIDList("foo", "bar")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFilterRender2(t *testing.T) {
	// Test adjacencies are removed for filtered nodes.
	renderer := render.Filter{
		FilterFunc: func(node report.Node) bool {
			return node.ID != "bar"
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode().WithID("foo").WithAdjacent("bar"),
			"bar": report.MakeNode().WithID("bar").WithAdjacent("foo"),
			"baz": report.MakeNode().WithID("baz"),
		}},
	}

	have := renderer.Render(report.MakeReport())
	if have["foo"].Adjacency.Contains("bar") {
		t.Error("adjacencies for removed nodes should have been removed")
	}
}

func TestFilterUnconnectedPseudoNodes(t *testing.T) {
	// Test pseudo nodes that are made unconnected by filtering
	// are also removed.
	{
		nodes := report.Nodes{
			"foo": report.MakeNode().WithID("foo").WithAdjacent("bar"),
			"bar": report.MakeNode().WithID("bar").WithAdjacent("baz"),
			"baz": report.MakeNode().WithID("baz").WithTopology(render.Pseudo),
		}
		renderer := render.Filter{
			FilterFunc: func(node report.Node) bool {
				return true
			},
			Renderer: mockRenderer{Nodes: nodes},
		}
		want := nodes
		have := renderer.Render(report.MakeReport())
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		renderer := render.Filter{
			FilterFunc: func(node report.Node) bool {
				return node.ID != "bar"
			},
			Renderer: mockRenderer{Nodes: report.Nodes{
				"foo": report.MakeNode().WithID("foo").WithAdjacent("bar"),
				"bar": report.MakeNode().WithID("bar").WithAdjacent("baz"),
				"baz": report.MakeNode().WithID("baz").WithTopology(render.Pseudo),
			}},
		}
		have := renderer.Render(report.MakeReport())
		if _, ok := have["baz"]; ok {
			t.Error("expected the unconnected pseudonode baz to have been removed")
		}
	}
	{
		renderer := render.Filter{
			FilterFunc: func(node report.Node) bool {
				return node.ID != "bar"
			},
			Renderer: mockRenderer{Nodes: report.Nodes{
				"foo": report.MakeNode().WithID("foo"),
				"bar": report.MakeNode().WithID("bar").WithAdjacent("foo"),
				"baz": report.MakeNode().WithID("baz").WithTopology(render.Pseudo).WithAdjacent("bar"),
			}},
		}
		have := renderer.Render(report.MakeReport())
		if _, ok := have["baz"]; ok {
			t.Error("expected the unconnected pseudonode baz to have been removed")
		}
	}
}

func TestFilterUnconnectedSelf(t *testing.T) {
	// Test nodes that are only connected to themselves are filtered.
	{
		nodes := report.Nodes{
			"foo": report.MakeNode().WithID("foo").WithAdjacent("foo"),
		}
		renderer := render.FilterUnconnected(mockRenderer{Nodes: nodes})
		have := renderer.Render(report.MakeReport())
		if len(have) > 0 {
			t.Error("expected node only connected to self to be removed")
		}
	}
}

func TestFilterPseudo(t *testing.T) {
	// Test pseudonodes are removed
	{
		nodes := report.Nodes{
			"foo": report.MakeNode().WithID("foo"),
			"bar": report.MakeNode().WithID("bar").WithTopology(render.Pseudo),
		}
		renderer := render.FilterPseudo(mockRenderer{Nodes: nodes})
		have := renderer.Render(report.MakeReport())
		if _, ok := have["bar"]; ok {
			t.Error("expected pseudonode to be removed")
		}
	}
}
