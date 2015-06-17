package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

type mockRenderer struct {
	render.RenderableNodes
	aggregateMetadata report.AggregateMetadata
}

func (m mockRenderer) Render(rpt report.Report) render.RenderableNodes {
	return m.RenderableNodes
}
func (m mockRenderer) AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return m.aggregateMetadata
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{RenderableNodes: render.RenderableNodes{"foo": {ID: "foo"}}},
		mockRenderer{RenderableNodes: render.RenderableNodes{"bar": {ID: "bar"}}},
	})

	want := render.RenderableNodes{"foo": {ID: "foo"}, "bar": {ID: "bar"}}
	have := renderer.Render(report.MakeReport())

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestReduceEdge(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"foo": 1}},
		mockRenderer{aggregateMetadata: report.AggregateMetadata{"bar": 2}},
	})

	want := report.AggregateMetadata{"foo": 1, "bar": 2}
	have := renderer.AggregateMetadata(report.MakeReport(), "", "")

	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender(t *testing.T) {
	// 1. Check when we return false, the node gets filtered out
	{
		mapper := render.Map{
			MapFunc: func(nodes render.RenderableNode) (render.RenderableNode, bool) {
				return render.RenderableNode{}, false
			},
			Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
				"foo": {ID: "foo"},
			}},
		}
		want := render.RenderableNodes{}
		have := mapper.Render(report.MakeReport())
		if !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}

	// 2. Check we can remap two nodes into one
	{
		mapper := render.Map{
			MapFunc: func(nodes render.RenderableNode) (render.RenderableNode, bool) {
				return render.RenderableNode{ID: "bar"}, true
			},
			Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
				"foo": {ID: "foo"},
				"baz": {ID: "baz"},
			}},
		}
		want := render.RenderableNodes{
			"bar": render.RenderableNode{ID: "bar"},
		}
		have := mapper.Render(report.MakeReport())
		if !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}

	// 3. Check we can remap adjacencies
	{
		mapper := render.Map{
			MapFunc: func(nodes render.RenderableNode) (render.RenderableNode, bool) {
				return render.RenderableNode{ID: "_" + nodes.ID}, true
			},
			Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
				"foo": {ID: "foo", Adjacency: report.MakeIDList("baz")},
				"baz": {ID: "baz", Adjacency: report.MakeIDList("foo")},
			}},
		}
		want := render.RenderableNodes{
			"_foo": {ID: "_foo", Adjacency: report.MakeIDList("_baz")},
			"_baz": {ID: "_baz", Adjacency: report.MakeIDList("_foo")},
		}
		have := mapper.Render(report.MakeReport())
		if !reflect.DeepEqual(want, have) {
			t.Errorf("want %+v, have %+v", want, have)
		}
	}
}

func TestMapEdge(t *testing.T) {
	selector := func(_ report.Report) report.Topology {
		return report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				"foo": report.NodeMetadata{"id": "foo"},
				"bar": report.NodeMetadata{"id": "bar"},
			},
			Adjacency: report.Adjacency{
				">foo": report.MakeIDList("bar"),
				">bar": report.MakeIDList("foo"),
			},
			EdgeMetadatas: report.EdgeMetadatas{
				"foo|bar": report.EdgeMetadata{WithBytes: true, BytesIngress: 1, BytesEgress: 2},
				"bar|foo": report.EdgeMetadata{WithBytes: true, BytesIngress: 3, BytesEgress: 4},
			},
		}
	}

	identity := func(nmd report.NodeMetadata) (render.RenderableNode, bool) {
		return render.NewRenderableNode(nmd["id"], "", "", "", nmd), true
	}

	mapper := render.Map{
		MapFunc: func(nodes render.RenderableNode) (render.RenderableNode, bool) {
			return render.RenderableNode{ID: "_" + nodes.ID}, true
		},
		Renderer: render.LeafMap{
			Selector: selector,
			Mapper:   identity,
			Pseudo:   nil,
		},
	}

	want := report.AggregateMetadata{
		report.KeyBytesIngress: 1,
		report.KeyBytesEgress:  2,
	}
	have := mapper.AggregateMetadata(report.MakeReport(), "_foo", "_bar")
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestFilterRender(t *testing.T) {
	renderer := render.FilterUnconnected{
		Renderer: mockRenderer{RenderableNodes: render.RenderableNodes{
			"foo": {ID: "foo", Adjacency: report.MakeIDList("bar")},
			"bar": {ID: "bar", Adjacency: report.MakeIDList("foo")},
			"baz": {ID: "baz", Adjacency: report.MakeIDList()},
		}},
	}
	want := render.RenderableNodes{
		"foo": {ID: "foo", Adjacency: report.MakeIDList("bar")},
		"bar": {ID: "bar", Adjacency: report.MakeIDList("foo")},
	}
	have := renderer.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}
