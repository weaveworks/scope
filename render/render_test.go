package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

type mockRenderer struct {
	render.RenderableNodes
}

func (m mockRenderer) Render(rpt report.Report) render.RenderableNodes {
	return m.RenderableNodes
}
func (m mockRenderer) Stats(rpt report.Report) render.Stats {
	return render.Stats{}
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{RenderableNodes: render.RenderableNodes{"foo": render.NewRenderableNode("foo")}},
		mockRenderer{RenderableNodes: render.RenderableNodes{"bar": render.NewRenderableNode("bar")}},
	})

	want := render.RenderableNodes{
		"foo": render.NewRenderableNode("foo"),
		"bar": render.NewRenderableNode("bar"),
	}
	have := renderer.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender1(t *testing.T) {
	// 1. Check when we return false, the node gets filtered out
	mapper := render.Map{
		MapFunc: func(nodes render.RenderableNode, _ report.Networks) render.RenderableNodes {
			return render.RenderableNodes{}
		},
		Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
			"foo": render.NewRenderableNode("foo"),
		}},
	}
	want := render.RenderableNodes{}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender2(t *testing.T) {
	// 2. Check we can remap two nodes into one
	mapper := render.Map{
		MapFunc: func(nodes render.RenderableNode, _ report.Networks) render.RenderableNodes {
			return render.RenderableNodes{
				"bar": render.NewRenderableNode("bar"),
			}
		},
		Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
			"foo": render.NewRenderableNode("foo"),
			"baz": render.NewRenderableNode("baz"),
		}},
	}
	want := render.RenderableNodes{
		"bar": render.NewRenderableNode("bar"),
	}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestMapRender3(t *testing.T) {
	// 3. Check we can remap adjacencies
	mapper := render.Map{
		MapFunc: func(nodes render.RenderableNode, _ report.Networks) render.RenderableNodes {
			id := "_" + nodes.ID
			return render.RenderableNodes{id: render.NewRenderableNode(id)}
		},
		Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
			"foo": render.NewRenderableNode("foo").WithNode(report.MakeNode().WithAdjacent("baz")),
			"baz": render.NewRenderableNode("baz").WithNode(report.MakeNode().WithAdjacent("foo")),
		}},
	}
	want := render.RenderableNodes{
		"_foo": render.NewRenderableNode("_foo").WithNode(report.MakeNode().WithAdjacent("_baz")),
		"_baz": render.NewRenderableNode("_baz").WithNode(report.MakeNode().WithAdjacent("_foo")),
	}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func newu64(value uint64) *uint64 { return &value }
