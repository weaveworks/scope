package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestFilterRender(t *testing.T) {
	renderer := mockRenderer{Nodes: report.Nodes{
		"foo": report.MakeNode("foo").WithAdjacent("bar"),
		"bar": report.MakeNode("bar").WithAdjacent("foo"),
		"baz": report.MakeNode("baz"),
	}}
	have := report.MakeIDList()
	for id := range renderer.Render(report.MakeReport(), render.FilterUnconnected).Nodes {
		have = have.Add(id)
	}
	want := report.MakeIDList("foo", "bar")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFilterRender2(t *testing.T) {
	// Test adjacencies are removed for filtered nodes.
	filter := func(renderer render.Renderer) render.Renderer {
		return &render.Filter{
			FilterFunc: func(node report.Node) bool {
				return node.ID != "bar"
			},
			Renderer: renderer,
		}
	}
	renderer := mockRenderer{Nodes: report.Nodes{
		"foo": report.MakeNode("foo").WithAdjacent("bar"),
		"bar": report.MakeNode("bar").WithAdjacent("foo"),
		"baz": report.MakeNode("baz"),
	}}

	have := renderer.Render(report.MakeReport(), filter).Nodes
	if have["foo"].Adjacency.Contains("bar") {
		t.Error("adjacencies for removed nodes should have been removed")
	}
}

func TestFilterUnconnectedPseudoNodes(t *testing.T) {
	// Test pseudo nodes that are made unconnected by filtering
	// are also removed.
	{
		nodes := report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("bar"),
			"bar": report.MakeNode("bar").WithAdjacent("baz"),
			"baz": report.MakeNode("baz").WithTopology(render.Pseudo),
		}
		renderer := mockRenderer{Nodes: nodes}
		filter := func(renderer render.Renderer) render.Renderer {
			return &render.Filter{
				FilterFunc: func(node report.Node) bool {
					return true
				},
				Renderer: renderer,
			}
		}
		want := nodes
		have := renderer.Render(report.MakeReport(), filter).Nodes
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		filter := func(renderer render.Renderer) render.Renderer {
			return &render.Filter{
				FilterFunc: func(node report.Node) bool {
					return node.ID != "bar"
				},
				Renderer: renderer,
			}
		}
		renderer := mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("bar"),
			"bar": report.MakeNode("bar").WithAdjacent("baz"),
			"baz": report.MakeNode("baz").WithTopology(render.Pseudo),
		}}
		have := renderer.Render(report.MakeReport(), filter).Nodes
		if _, ok := have["baz"]; ok {
			t.Error("expected the unconnected pseudonode baz to have been removed")
		}
	}
	{
		filter := func(renderer render.Renderer) render.Renderer {
			return &render.Filter{
				FilterFunc: func(node report.Node) bool {
					return node.ID != "bar"
				},
				Renderer: renderer,
			}
		}
		renderer := mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
			"bar": report.MakeNode("bar").WithAdjacent("foo"),
			"baz": report.MakeNode("baz").WithTopology(render.Pseudo).WithAdjacent("bar"),
		}}
		have := renderer.Render(report.MakeReport(), filter).Nodes
		if _, ok := have["baz"]; ok {
			t.Error("expected the unconnected pseudonode baz to have been removed")
		}
	}
}

func TestFilterUnconnectedSelf(t *testing.T) {
	// Test nodes that are only connected to themselves are filtered.
	{
		nodes := report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("foo"),
		}
		renderer := mockRenderer{Nodes: nodes}
		have := renderer.Render(report.MakeReport(), render.FilterUnconnected).Nodes
		if len(have) > 0 {
			t.Error("expected node only connected to self to be removed")
		}
	}
}
