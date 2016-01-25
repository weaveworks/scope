package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestMergeRenderableNodes(t *testing.T) {
	nodes1 := render.MakeRenderableNodes(
		render.NewRenderableNode("foo"),
		render.NewRenderableNode("bar"),
	)
	nodes2 := render.MakeRenderableNodes(
		render.NewRenderableNode("bar"),
		render.NewRenderableNode("baz"),
	)
	want := (render.MakeRenderableNodes(
		render.NewRenderableNode("foo"),
		render.NewRenderableNode("bar"),
		render.NewRenderableNode("baz"),
	)).Prune()
	have := nodes1.Merge(nodes2).Prune()
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
		Children:   report.MakeNodeSet(report.MakeNode().WithID("child1")),
	}
	node2 := render.RenderableNode{
		ID:         "foo",
		LabelMajor: "major",
		LabelMinor: "",
		Rank:       "rank",
		Pseudo:     false,
		Node:       report.MakeNode().WithAdjacent("a2"),
		Children:   report.MakeNodeSet(report.MakeNode().WithID("child2")),
	}
	want := render.RenderableNode{
		ID:           "foo",
		LabelMajor:   "major",
		LabelMinor:   "minor",
		Rank:         "rank",
		Pseudo:       false,
		Node:         report.MakeNode().WithID("foo").WithAdjacent("a1").WithAdjacent("a2"),
		Children:     report.MakeNodeSet(report.MakeNode().WithID("child1"), report.MakeNode().WithID("child2")),
		EdgeMetadata: report.EdgeMetadata{},
	}.Prune()
	have := node1.Merge(node2).Prune()
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestRenderableNodesDeepEqual(t *testing.T) {
	for _, c := range []struct {
		name string
		a, b interface{}
		want bool
	}{
		{
			name: "zero values",
			a:    render.RenderableNodes{},
			b:    render.RenderableNodes{},
			want: true,
		},
		{
			name: "mismatched types",
			a:    render.RenderableNodes{},
			b:    int(5),
			want: false,
		},
		{
			name: "nil argument",
			a:    render.EmptyRenderableNodes,
			b:    (*render.RenderableNodes)(nil),
			want: false,
		},
		{
			name: "nil receiver",
			a:    (*render.RenderableNodes)(nil),
			b:    render.EmptyRenderableNodes,
			want: false,
		},
		/*
			{
				name: "both nil",
				a:    (*render.RenderableNodes)(nil),
				b:    (*render.RenderableNodes)(nil),
				want: true,
			},
		*/
		{
			name: "plain empty sets",
			a:    render.EmptyRenderableNodes,
			b:    render.EmptyRenderableNodes,
			want: true,
		},
		{
			name: "one set with node(s)",
			a: render.MakeRenderableNodes(render.RenderableNode{
				ID:         "foo",
				LabelMinor: "minor",
				Pseudo:     false,
				Node:       report.MakeNode().WithAdjacent("a1"),
				Children:   report.MakeNodeSet(report.MakeNode().WithID("child1")),
			}),
			b:    render.EmptyRenderableNodes,
			want: false,
		},
		{
			name: "matching sets with node(s)",
			a: render.MakeRenderableNodes(render.RenderableNode{
				ID:         "foo",
				LabelMinor: "minor",
				Pseudo:     false,
				Node:       report.MakeNode().WithAdjacent("a1"),
				Children:   report.MakeNodeSet(report.MakeNode().WithID("child1")),
			}),
			b: render.MakeRenderableNodes(render.RenderableNode{
				ID:         "foo",
				LabelMinor: "minor",
				Pseudo:     false,
				Node:       report.MakeNode().WithAdjacent("a1"),
				Children:   report.MakeNodeSet(report.MakeNode().WithID("child1")),
			}),
			want: true,
		},
	} {
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("[%s] %#v.DeepEqual(%#v) panic: %v", c.name, c.a, c.b, r)
				}
			}()
			if reflect.DeepEqual(c.a, c.b) != c.want {
				return fmt.Errorf("%s\t%#v.DeepEqual(%#v) != %v", c.name, c.a, c.b, c.want)
			}
			return nil
		}()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestRenderableNodesKeys(t *testing.T) {
	for _, c := range []struct {
		name string
		a    render.RenderableNodes
		want []string
	}{
		{
			name: "zero values",
			a:    render.RenderableNodes{},
			want: nil,
		},
		{
			name: "empty",
			a:    render.EmptyRenderableNodes,
			want: []string{},
		},
		{
			name: "one node",
			a:    render.MakeRenderableNodes(render.NewRenderableNode("a")),
			want: []string{"a"},
		},
		{
			name: "three nodes (out of order)",
			a: render.MakeRenderableNodes(
				render.NewRenderableNode("b"),
				render.NewRenderableNode("a"),
				render.NewRenderableNode("c"),
			),
			want: []string{"a", "b", "c"},
		},
	} {
		if have := c.a.Keys(); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s\t%#v.Keys() expected: %#v, got: %#v", c.name, c.a, c.want, have)
		}
	}
}
