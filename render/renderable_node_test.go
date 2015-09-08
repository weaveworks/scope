package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestMergeRenderableNodes(t *testing.T) {
	nodes1 := render.RenderableNodes{
		"foo": render.NewRenderableNode("foo"),
		"bar": render.NewRenderableNode("bar"),
	}
	nodes2 := render.RenderableNodes{
		"bar": render.NewRenderableNode("bar"),
		"baz": render.NewRenderableNode("baz"),
	}
	want := (render.RenderableNodes{
		"foo": render.NewRenderableNode("foo"),
		"bar": render.NewRenderableNode("bar"),
		"baz": render.NewRenderableNode("baz"),
	}).Prune()
	have := nodes1.Merge(nodes2)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestMergeRenderableNode(t *testing.T) {
	node1 := render.RenderableNode{
		ID:         "foo",
		LabelMajor: "",
		LabelMinor: "minor",
		Rank:       "",
		Pseudo:     false,
		Node:       report.MakeNode().WithAdjacent("a1"),
		Origins:    report.MakeIDList("o1"),
	}
	node2 := render.RenderableNode{
		ID:         "foo",
		LabelMajor: "major",
		LabelMinor: "",
		Rank:       "rank",
		Pseudo:     false,
		Node:       report.MakeNode().WithAdjacent("a2"),
		Origins:    report.MakeIDList("o2"),
	}
	want := render.RenderableNode{
		ID:           "foo",
		LabelMajor:   "major",
		LabelMinor:   "minor",
		Rank:         "rank",
		Pseudo:       false,
		Node:         report.MakeNode().WithAdjacency(report.MakeIDList("a1", "a2")),
		Origins:      report.MakeIDList("o1", "o2"),
		EdgeMetadata: report.EdgeMetadata{},
	}
	have := node1.Merge(node2)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
