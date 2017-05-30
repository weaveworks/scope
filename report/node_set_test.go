package report_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

var benchmarkResult report.NodeSet

type nodeSpec struct {
	id string
}

func TestMakeNodeSet(t *testing.T) {
	for _, testcase := range []struct {
		inputs []string
		wants  []string
	}{
		{inputs: nil, wants: nil},
		{
			inputs: []string{"a"},
			wants:  []string{"a"},
		},
		{
			inputs: []string{"b", "c", "a"},
			wants:  []string{"a", "b", "c"},
		},
		{
			inputs: []string{"a", "a", "a"},
			wants:  []string{"a"},
		},
	} {
		var inputs []report.Node
		for _, id := range testcase.inputs {
			inputs = append(inputs, report.MakeNode(id))
		}
		set := report.MakeNodeSet(inputs...)
		var have []string
		set.ForEach(func(node report.Node) { have = append(have, node.ID) })
		sort.Strings(have)
		if !reflect.DeepEqual(testcase.wants, have) {
			t.Errorf("%#v: want %#v, have %#v", testcase.inputs, testcase.wants, have)
		}
	}
}

func BenchmarkMakeNodeSet(b *testing.B) {
	nodes := []report.Node{}
	for i := 1000; i >= 0; i-- {
		nodes = append(nodes, report.MakeNodeWith(fmt.Sprint(i), map[string]string{
			"a": "1",
			"b": "2",
		}))
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = report.MakeNodeSet(nodes...)
	}
}

func TestNodeSetAdd(t *testing.T) {
	for _, testcase := range []struct {
		input report.NodeSet
		nodes []report.Node
		want  report.NodeSet
	}{
		{
			input: report.NodeSet{},
			nodes: []report.Node{},
			want:  report.NodeSet{},
		},
		{
			input: report.EmptyNodeSet,
			nodes: []report.Node{},
			want:  report.EmptyNodeSet,
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []report.Node{},
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.EmptyNodeSet,
			nodes: []report.Node{report.MakeNode("a")},
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []report.Node{report.MakeNode("a")},
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("b")),
			nodes: []report.Node{
				report.MakeNode("a"),
				report.MakeNode("b"),
			},
			want: report.MakeNodeSet(
				report.MakeNode("a"),
				report.MakeNode("b"),
			),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []report.Node{
				report.MakeNode("c"),
				report.MakeNode("b"),
			},
			want: report.MakeNodeSet(
				report.MakeNode("a"),
				report.MakeNode("b"),
				report.MakeNode("c"),
			),
		},
		{
			input: report.MakeNodeSet(
				report.MakeNode("a"),
				report.MakeNode("c"),
			),
			nodes: []report.Node{
				report.MakeNode("b"),
				report.MakeNode("b"),
				report.MakeNode("b"),
			},
			want: report.MakeNodeSet(
				report.MakeNode("a"),
				report.MakeNode("b"),
				report.MakeNode("c"),
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

func BenchmarkNodeSetAdd(b *testing.B) {
	n := report.EmptyNodeSet
	for i := 0; i < 600; i++ {
		n = n.Add(
			report.MakeNodeWith(fmt.Sprint(i), map[string]string{
				"a": "1",
				"b": "2",
			}),
		)
	}

	node := report.MakeNodeWith("401.5", map[string]string{
		"a": "1",
		"b": "2",
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Add(node)
	}
}

func TestNodeSetDelete(t *testing.T) {
	for _, testcase := range []struct {
		input report.NodeSet
		nodes []string
		want  report.NodeSet
	}{
		{
			input: report.NodeSet{},
			nodes: []string{},
			want:  report.NodeSet{},
		},
		{
			input: report.EmptyNodeSet,
			nodes: []string{},
			want:  report.EmptyNodeSet,
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []string{},
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.EmptyNodeSet,
			nodes: []string{"a"},
			want:  report.EmptyNodeSet,
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []string{"a"},
			want:  report.EmptyNodeSet,
		},
		{
			input: report.MakeNodeSet(report.MakeNode("b")),
			nodes: []string{"a", "b"},
			want:  report.EmptyNodeSet,
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			nodes: []string{"c", "b"},
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("c")),
			nodes: []string{"a", "a", "a"},
			want:  report.MakeNodeSet(report.MakeNode("c")),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Delete(testcase.nodes...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.nodes, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.nodes)
		}
	}
}

func TestNodeSetMerge(t *testing.T) {
	for _, testcase := range []struct {
		input report.NodeSet
		other report.NodeSet
		want  report.NodeSet
	}{
		{input: report.NodeSet{}, other: report.NodeSet{}, want: report.NodeSet{}},
		{input: report.EmptyNodeSet, other: report.EmptyNodeSet, want: report.EmptyNodeSet},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			other: report.EmptyNodeSet,
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.EmptyNodeSet,
			other: report.MakeNodeSet(report.MakeNode("a")),
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			other: report.MakeNodeSet(report.MakeNode("b")),
			want:  report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("b")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("b")),
			other: report.MakeNodeSet(report.MakeNode("a")),
			want:  report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("b")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a")),
			other: report.MakeNodeSet(report.MakeNode("a")),
			want:  report.MakeNodeSet(report.MakeNode("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("c")),
			other: report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("b")),
			want:  report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("b"), report.MakeNode("c")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode("b")),
			other: report.MakeNodeSet(report.MakeNode("a")),
			want:  report.MakeNodeSet(report.MakeNode("a"), report.MakeNode("b")),
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

func BenchmarkNodeSetMerge(b *testing.B) {
	n, other := report.NodeSet{}, report.NodeSet{}
	for i := 0; i < 600; i++ {
		n = n.Add(
			report.MakeNodeWith(fmt.Sprint(i), map[string]string{
				"a": "1",
				"b": "2",
			}),
		)
	}

	for i := 400; i < 1000; i++ {
		other = other.Add(
			report.MakeNodeWith(fmt.Sprint(i), map[string]string{
				"c": "1",
				"d": "2",
			}),
		)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Merge(other)
	}
}
