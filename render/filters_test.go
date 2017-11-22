package render_test

import (
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func isNotBar(node report.Node) bool {
	return node.ID != "bar"
}

func TestFilterRender(t *testing.T) {
	renderer := mockRenderer{Nodes: report.Nodes{
		"foo": report.MakeNode("foo").WithAdjacent("bar"),
		"bar": report.MakeNode("bar").WithAdjacent("foo"),
		"baz": report.MakeNode("baz"),
	}}
	have := report.MakeIDList()
	for id := range render.Render(report.MakeReport(), render.ColorConnected(renderer), render.FilterFunc(render.IsConnected)).Nodes {
		have = have.Add(id)
	}
	want := report.MakeIDList("foo", "bar")
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFilterRender2(t *testing.T) {
	// Test adjacencies are removed for filtered nodes.
	renderer := mockRenderer{Nodes: report.Nodes{
		"foo": report.MakeNode("foo").WithAdjacent("bar"),
		"bar": report.MakeNode("bar").WithAdjacent("foo"),
		"baz": report.MakeNode("baz"),
	}}
	have := render.Render(report.MakeReport(), renderer, render.FilterFunc(isNotBar)).Nodes
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
		want := nodes
		have := render.Render(report.MakeReport(), renderer, render.Transformers(nil)).Nodes
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		renderer := mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("bar"),
			"bar": report.MakeNode("bar").WithAdjacent("baz"),
			"baz": report.MakeNode("baz").WithTopology(render.Pseudo),
		}}
		have := render.Render(report.MakeReport(), renderer, render.FilterFunc(isNotBar)).Nodes
		if _, ok := have["baz"]; ok {
			t.Error("expected the unconnected pseudonode baz to have been removed")
		}
	}
	{
		renderer := mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
			"bar": report.MakeNode("bar").WithAdjacent("foo"),
			"baz": report.MakeNode("baz").WithTopology(render.Pseudo).WithAdjacent("bar"),
		}}
		have := render.Render(report.MakeReport(), renderer, render.FilterFunc(isNotBar)).Nodes
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
		have := render.Render(report.MakeReport(), render.ColorConnected(renderer), render.FilterFunc(render.IsConnected)).Nodes
		if len(have) > 0 {
			t.Error("expected node only connected to self to be removed")
		}
	}
}
