package render_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

var benchmarkResult render.RenderableNodeSet

type nodeSpec struct {
	topology string
	id       string
}

func renderableNode(n report.Node) render.RenderableNode {
	node := render.NewRenderableNode(n.ID)
	node.Topology = n.Topology
	return node
}

func TestMakeRenderableNodeSet(t *testing.T) {
	for _, testcase := range []struct {
		inputs []nodeSpec
		wants  []nodeSpec
	}{
		{inputs: nil, wants: nil},
		{inputs: []nodeSpec{}, wants: []nodeSpec{}},
		{
			inputs: []nodeSpec{{"", "a"}},
			wants:  []nodeSpec{{"", "a"}},
		},
		{
			inputs: []nodeSpec{{"", "a"}, {"", "a"}, {"1", "a"}},
			wants:  []nodeSpec{{"", "a"}, {"1", "a"}},
		},
		{
			inputs: []nodeSpec{{"", "b"}, {"", "c"}, {"", "a"}},
			wants:  []nodeSpec{{"", "a"}, {"", "b"}, {"", "c"}},
		},
		{
			inputs: []nodeSpec{{"2", "a"}, {"3", "a"}, {"1", "a"}},
			wants:  []nodeSpec{{"1", "a"}, {"2", "a"}, {"3", "a"}},
		},
	} {
		var (
			inputs []render.RenderableNode
			wants  []render.RenderableNode
		)
		for _, spec := range testcase.inputs {
			node := render.NewRenderableNode(spec.id)
			node.Topology = spec.topology
			inputs = append(inputs, node)
		}
		for _, spec := range testcase.wants {
			node := render.NewRenderableNode(spec.id)
			node.Topology = spec.topology
			wants = append(wants, node)
		}
		if want, have := render.MakeRenderableNodeSet(wants...), render.MakeRenderableNodeSet(inputs...); !reflect.DeepEqual(want, have) {
			t.Errorf("%#v: want %#v, have %#v", inputs, wants, have)
		}
	}
}

func BenchmarkMakeRenderableNodeSet(b *testing.B) {
	nodes := []render.RenderableNode{}
	for i := 1000; i >= 0; i-- {
		node := report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
			"a": "1",
			"b": "2",
		})
		rn := render.NewRenderableNode(node.ID)
		rn.Node = node
		nodes = append(nodes, rn)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = render.MakeRenderableNodeSet(nodes...)
	}
}

func TestRenderableNodeSetAdd(t *testing.T) {
	for _, testcase := range []struct {
		input render.RenderableNodeSet
		nodes []render.RenderableNode
		want  render.RenderableNodeSet
	}{
		{
			input: render.RenderableNodeSet{},
			nodes: []render.RenderableNode{},
			want:  render.RenderableNodeSet{},
		},
		{
			input: render.EmptyRenderableNodeSet,
			nodes: []render.RenderableNode{},
			want:  render.EmptyRenderableNodeSet,
		},
		{
			input: render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
			nodes: []render.RenderableNode{},
			want:  render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
		},
		{
			input: render.EmptyRenderableNodeSet,
			nodes: []render.RenderableNode{renderableNode(report.MakeNode().WithID("a"))},
			want:  render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
		},
		{
			input: render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
			nodes: []render.RenderableNode{renderableNode(report.MakeNode().WithID("a"))},
			want:  render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
		},
		{
			input: render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("b"))),
			nodes: []render.RenderableNode{
				renderableNode(report.MakeNode().WithID("a")),
				renderableNode(report.MakeNode().WithID("b")),
			},
			want: render.MakeRenderableNodeSet(
				renderableNode(report.MakeNode().WithID("a")),
				renderableNode(report.MakeNode().WithID("b")),
			),
		},
		{
			input: render.MakeRenderableNodeSet(renderableNode(report.MakeNode().WithID("a"))),
			nodes: []render.RenderableNode{
				renderableNode(report.MakeNode().WithID("c")),
				renderableNode(report.MakeNode().WithID("b")),
			},
			want: render.MakeRenderableNodeSet(
				renderableNode(report.MakeNode().WithID("a")),
				renderableNode(report.MakeNode().WithID("b")),
				renderableNode(report.MakeNode().WithID("c")),
			),
		},
		{
			input: render.MakeRenderableNodeSet(
				renderableNode(report.MakeNode().WithID("a")),
				renderableNode(report.MakeNode().WithID("c")),
			),
			nodes: []render.RenderableNode{
				renderableNode(report.MakeNode().WithID("b")),
				renderableNode(report.MakeNode().WithID("b")),
				renderableNode(report.MakeNode().WithID("b")),
			},
			want: render.MakeRenderableNodeSet(
				renderableNode(report.MakeNode().WithID("a")),
				renderableNode(report.MakeNode().WithID("b")),
				renderableNode(report.MakeNode().WithID("c")),
			),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Add(testcase.nodes...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.nodes, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.nodes)
		}
	}
}

func BenchmarkRenderableNodeSetAdd(b *testing.B) {
	n := render.EmptyRenderableNodeSet
	for i := 0; i < 600; i++ {
		n = n.Add(
			renderableNode(report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
				"a": "1",
				"b": "2",
			})),
		)
	}

	node := renderableNode(report.MakeNode().WithID("401.5").WithLatests(map[string]string{
		"a": "1",
		"b": "2",
	}))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Add(node)
	}
}

func TestRenderableNodeSetMerge(t *testing.T) {
	for _, testcase := range []struct {
		input render.RenderableNodeSet
		other render.RenderableNodeSet
		want  render.RenderableNodeSet
	}{
		{input: render.RenderableNodeSet{}, other: render.RenderableNodeSet{}, want: render.RenderableNodeSet{}},
		{input: render.EmptyRenderableNodeSet, other: render.EmptyRenderableNodeSet, want: render.EmptyRenderableNodeSet},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			other: render.EmptyRenderableNodeSet,
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
		},
		{
			input: render.EmptyRenderableNodeSet,
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
		},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("b")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("b")),
		},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("b")),
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("b")),
		},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
		},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("c")),
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("b")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("b"), render.NewRenderableNode("c")),
		},
		{
			input: render.MakeRenderableNodeSet(render.NewRenderableNode("b")),
			other: render.MakeRenderableNodeSet(render.NewRenderableNode("a")),
			want:  render.MakeRenderableNodeSet(render.NewRenderableNode("a"), render.NewRenderableNode("b")),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Merge(testcase.other); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.other)
		}
	}
}

func BenchmarkRenderableNodeSetMerge(b *testing.B) {
	n, other := render.RenderableNodeSet{}, render.RenderableNodeSet{}
	for i := 0; i < 600; i++ {
		n = n.Add(
			renderableNode(report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
				"a": "1",
				"b": "2",
			})),
		)
	}

	for i := 400; i < 1000; i++ {
		other = other.Add(
			renderableNode(report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
				"c": "1",
				"d": "2",
			})),
		)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Merge(other)
	}
}
