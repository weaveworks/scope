package render_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestMergeRenderableNodes(t *testing.T) {
	nodes1 := render.RenderableNodes{
		"foo": render.RenderableNode{ID: "foo"},
		"bar": render.RenderableNode{ID: "bar"},
	}
	nodes2 := render.RenderableNodes{
		"bar": render.RenderableNode{ID: "bar"},
		"baz": render.RenderableNode{ID: "baz"},
	}
	want := expected.Sterilize(render.RenderableNodes{
		"foo": render.RenderableNode{ID: "foo"},
		"bar": render.RenderableNode{ID: "bar"},
		"baz": render.RenderableNode{ID: "baz"},
	})
	nodes1.Merge(nodes2)
	if have := expected.Sterilize(nodes1); !reflect.DeepEqual(want, have) {
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
		Adjacency:  report.MakeIDList("a1"),
		Origins:    report.MakeIDList("o1"),
	}
	node2 := render.RenderableNode{
		ID:         "foo",
		LabelMajor: "major",
		LabelMinor: "",
		Rank:       "rank",
		Pseudo:     false,
		Adjacency:  report.MakeIDList("a2"),
		Origins:    report.MakeIDList("o2"),
	}
	want := render.RenderableNode{
		ID:           "foo",
		LabelMajor:   "major",
		LabelMinor:   "minor",
		Rank:         "rank",
		Pseudo:       false,
		Adjacency:    report.MakeIDList("a1", "a2"),
		Origins:      report.MakeIDList("o1", "o2"),
		NodeMetadata: report.MakeNodeMetadata(),
		EdgeMetadata: report.EdgeMetadata{},
	}
	node1.Merge(node2)
	if have := node1; !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
