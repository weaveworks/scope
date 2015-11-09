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
	edgeMetadata report.EdgeMetadata
}

func (m mockRenderer) Render(rpt report.Report) render.RenderableNodes {
	return m.RenderableNodes
}
func (m mockRenderer) EdgeMetadata(rpt report.Report, localID, remoteID string) report.EdgeMetadata {
	return m.edgeMetadata
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

func TestReduceEdge(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{edgeMetadata: report.EdgeMetadata{EgressPacketCount: newu64(1)}},
		mockRenderer{edgeMetadata: report.EdgeMetadata{EgressPacketCount: newu64(2)}},
	})

	want := report.EdgeMetadata{EgressPacketCount: newu64(3)}
	have := renderer.EdgeMetadata(report.MakeReport(), "", "")
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

func TestMapEdge(t *testing.T) {
	selector := render.TopologySelector(func(_ report.Report) render.RenderableNodes {
		return render.MakeRenderableNodes(report.Topology{
			Nodes: report.Nodes{
				"foo": report.MakeNode().WithMetadata(map[string]string{
					"id": "foo",
				}).WithEdge("bar", report.EdgeMetadata{
					EgressPacketCount: newu64(1),
					EgressByteCount:   newu64(2),
				}),

				"bar": report.MakeNode().WithMetadata(map[string]string{
					"id": "bar",
				}).WithEdge("foo", report.EdgeMetadata{
					EgressPacketCount: newu64(3),
					EgressByteCount:   newu64(4),
				}),
			},
		})
	})

	mapper := render.Map{
		MapFunc: func(node render.RenderableNode, _ report.Networks) render.RenderableNodes {
			id := "_" + node.ID
			return render.RenderableNodes{id: render.NewDerivedNode(id, node)}
		},
		Renderer: selector,
	}

	have := mapper.Render(report.MakeReport()).Prune()
	want := (render.RenderableNodes{
		"_foo": {
			ID:      "_foo",
			Origins: report.MakeIDList("foo"),
			Node:    report.MakeNode().WithAdjacent("_bar"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount:  newu64(1),
				EgressByteCount:    newu64(2),
				IngressPacketCount: newu64(3),
				IngressByteCount:   newu64(4),
			},
		},
		"_bar": {
			ID:      "_bar",
			Origins: report.MakeIDList("bar"),
			Node:    report.MakeNode().WithAdjacent("_foo"),
			EdgeMetadata: report.EdgeMetadata{
				EgressPacketCount:  newu64(3),
				EgressByteCount:    newu64(4),
				IngressPacketCount: newu64(1),
				IngressByteCount:   newu64(2),
			},
		},
	}).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	if want, have := (report.EdgeMetadata{
		EgressPacketCount: newu64(1),
		EgressByteCount:   newu64(2),
	}), mapper.EdgeMetadata(report.MakeReport(), "_foo", "_bar"); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func newu64(value uint64) *uint64 { return &value }
