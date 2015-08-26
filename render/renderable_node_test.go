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
		"foo": render.RenderableNode{ID: "foo"},
		"bar": render.RenderableNode{ID: "bar"},
	}
	nodes2 := render.RenderableNodes{
		"bar": render.RenderableNode{ID: "bar"},
		"baz": render.RenderableNode{ID: "baz"},
	}
	want := sterilize(render.RenderableNodes{
		"foo": render.RenderableNode{ID: "foo"},
		"bar": render.RenderableNode{ID: "bar"},
		"baz": render.RenderableNode{ID: "baz"},
	}, false)
	nodes1.Merge(nodes2)
	if have := sterilize(nodes1, false); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func sterilize(r render.RenderableNodes, destructive bool) render.RenderableNodes {
	// Since introducing new map fields to the report.NodeMetadata type, its
	// zero value is •not valid• -- every time you need one, you need to use
	// the report.MakeNodeMetadata constructor. (Similarly, but not exactly
	// the same, is that a zero-value Adjacency is not the same as a created
	// but empty Adjacency.)
	//
	// But we're not doing this in tests. So this function sterilizes invalid
	// RenderableNodes by fixing all nil Metadata fields. The proper fix
	// involves lots of annoying changes to instantiation.
	//
	// The extra destructive parameter is to support a historical test use
	// case where we explicitly don't compare node metadata.
	for id, n := range r {
		if n.Adjacency == nil {
			n.Adjacency = report.IDList{}
		}
		if destructive || n.NodeMetadata.Metadata == nil {
			n.NodeMetadata.Metadata = map[string]string{}
		}
		if destructive || n.NodeMetadata.Counters == nil {
			n.NodeMetadata.Counters = map[string]int{}
		}
		r[id] = n
	}
	return r
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
