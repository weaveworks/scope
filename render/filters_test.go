package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestFilterRender(t *testing.T) {
	renderer := render.FilterUnconnected(
		mockRenderer{RenderableNodes: render.MakeRenderableNodes(
			render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("bar")},
			render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("foo")},
			render.RenderableNode{ID: "baz", Node: report.MakeNode()},
		)})
	want := render.MakeRenderableNodes(
		render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("bar")},
		render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("foo")},
	)
	have := renderer.Render(report.MakeReport()).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFilterRender2(t *testing.T) {
	// Test adjacencies are removed for filtered nodes.
	renderer := render.Filter{
		FilterFunc: func(node render.RenderableNode) bool {
			return node.ID != "bar"
		},
		Renderer: mockRenderer{RenderableNodes: render.MakeRenderableNodes(
			render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("bar")},
			render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("foo")},
			render.RenderableNode{ID: "baz", Node: report.MakeNode()},
		)},
	}
	want := render.MakeRenderableNodes(
		render.RenderableNode{ID: "foo", Node: report.MakeNode()},
		render.RenderableNode{ID: "baz", Node: report.MakeNode()},
	)
	have := renderer.Render(report.MakeReport()).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFilterUnconnectedPseudoNodes(t *testing.T) {
	// Test pseudo nodes that are made unconnected by filtering
	// are also removed.
	{
		nodes := render.MakeRenderableNodes(
			render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("bar")},
			render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("baz")},
			render.RenderableNode{ID: "baz", Node: report.MakeNode(), Pseudo: true},
		)
		renderer := render.Filter{
			FilterFunc: func(node render.RenderableNode) bool {
				return true
			},
			Renderer: mockRenderer{RenderableNodes: nodes},
		}
		want := nodes.Prune()
		have := renderer.Render(report.MakeReport()).Prune()
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		renderer := render.Filter{
			FilterFunc: func(node render.RenderableNode) bool {
				return node.ID != "bar"
			},
			Renderer: mockRenderer{RenderableNodes: render.MakeRenderableNodes(
				render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("bar")},
				render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("baz")},
				render.RenderableNode{ID: "baz", Node: report.MakeNode(), Pseudo: true},
			)},
		}
		want := render.MakeRenderableNodes(render.RenderableNode{ID: "foo", Node: report.MakeNode()})
		have := renderer.Render(report.MakeReport()).Prune()
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
	{
		renderer := render.Filter{
			FilterFunc: func(node render.RenderableNode) bool {
				return node.ID != "bar"
			},
			Renderer: mockRenderer{RenderableNodes: render.MakeRenderableNodes(
				render.RenderableNode{ID: "foo", Node: report.MakeNode()},
				render.RenderableNode{ID: "bar", Node: report.MakeNode().WithAdjacent("foo")},
				render.RenderableNode{ID: "baz", Node: report.MakeNode().WithAdjacent("bar"), Pseudo: true},
			)},
		}
		want := render.MakeRenderableNodes(render.RenderableNode{ID: "foo", Node: report.MakeNode()})
		have := renderer.Render(report.MakeReport()).Prune()
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestFilterUnconnectedSelf(t *testing.T) {
	// Test nodes that are only connected to themselves are filtered.
	{
		nodes := render.MakeRenderableNodes(render.RenderableNode{ID: "foo", Node: report.MakeNode().WithAdjacent("foo")})
		renderer := render.FilterUnconnected(mockRenderer{RenderableNodes: nodes})
		want := render.MakeRenderableNodes()
		have := renderer.Render(report.MakeReport()).Prune()
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestFilterPseudo(t *testing.T) {
	// Test pseudonodes are removed
	{
		nodes := render.MakeRenderableNodes(
			render.RenderableNode{ID: "foo", Node: report.MakeNode()},
			render.RenderableNode{ID: "bar", Pseudo: true, Node: report.MakeNode()},
		)
		renderer := render.FilterPseudo(mockRenderer{RenderableNodes: nodes})
		want := render.MakeRenderableNodes(render.RenderableNode{ID: "foo", Node: report.MakeNode()})
		have := renderer.Render(report.MakeReport()).Prune()
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}
