package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

type mockRenderer struct {
	report.Nodes
}

func (m mockRenderer) Render(rpt report.Report, d render.Decorator) render.Nodes {
	if d != nil {
		return d(mockRenderer{m.Nodes}).Render(rpt, nil)
	}
	return render.Nodes{Nodes: m.Nodes}
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{Nodes: report.Nodes{"foo": report.MakeNode("foo")}},
		mockRenderer{Nodes: report.Nodes{"bar": report.MakeNode("bar")}},
	})

	want := report.Nodes{
		"foo": report.MakeNode("foo"),
		"bar": report.MakeNode("bar"),
	}
	have := renderer.Render(report.MakeReport(), FilterNoop).Nodes
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender1(t *testing.T) {
	// 1. Check when we return false, the node gets filtered out
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			return report.Nodes{}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
		}},
	}
	want := report.Nodes{}
	have := mapper.Render(report.MakeReport(), FilterNoop).Nodes
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender2(t *testing.T) {
	// 2. Check we can remap two nodes into one
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			return report.Nodes{
				"bar": report.MakeNode("bar"),
			}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
			"baz": report.MakeNode("baz"),
		}},
	}
	want := report.Nodes{
		"bar": report.MakeNode("bar"),
	}
	have := mapper.Render(report.MakeReport(), FilterNoop).Nodes
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestMapRender3(t *testing.T) {
	// 3. Check we can remap adjacencies
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			id := "_" + nodes.ID
			return report.Nodes{id: report.MakeNode(id)}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("baz"),
			"baz": report.MakeNode("baz").WithAdjacent("foo"),
		}},
	}
	want := report.Nodes{
		"_foo": report.MakeNode("_foo").WithAdjacent("_baz"),
		"_baz": report.MakeNode("_baz").WithAdjacent("_foo"),
	}
	have := mapper.Render(report.MakeReport(), FilterNoop).Nodes
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func newu64(value uint64) *uint64 { return &value }
