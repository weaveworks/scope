package report_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

var benchmarkResult report.NodeSet

type nodeSpec struct {
	topology string
	id       string
}

func TestMakeNodeSet(t *testing.T) {
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
			inputs []report.Node
			wants  []report.Node
		)
		for _, spec := range testcase.inputs {
			inputs = append(inputs, report.MakeNode().WithTopology(spec.topology).WithID(spec.id))
		}
		for _, spec := range testcase.wants {
			wants = append(wants, report.MakeNode().WithTopology(spec.topology).WithID(spec.id))
		}
		if want, have := report.NodeSet(wants), report.MakeNodeSet(inputs...); !reflect.DeepEqual(want, have) {
			t.Errorf("%#v: want %#v, have %#v", inputs, wants, have)
		}
	}
}

func BenchmarkMakeNodeSet(b *testing.B) {
	nodes := []report.Node{}
	for i := 1000; i >= 0; i-- {
		node := report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
			"a": "1",
			"b": "2",
		})
		nodes = append(nodes, node)
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
		{input: report.NodeSet(nil), nodes: []report.Node{}, want: report.NodeSet(nil)},
		{
			input: report.MakeNodeSet(),
			nodes: []report.Node{},
			want:  report.MakeNodeSet(),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			nodes: []report.Node{},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(),
			nodes: []report.Node{report.MakeNode().WithID("a")},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			nodes: []report.Node{report.MakeNode().WithID("a")},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("b")),
			nodes: []report.Node{report.MakeNode().WithID("a"), report.MakeNode().WithID("b")},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			nodes: []report.Node{report.MakeNode().WithID("c"), report.MakeNode().WithID("b")},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b"), report.MakeNode().WithID("c")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("c")),
			nodes: []report.Node{report.MakeNode().WithID("b"), report.MakeNode().WithID("b"), report.MakeNode().WithID("b")},
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b"), report.MakeNode().WithID("c")),
		},
	} {
		originalLen := len(testcase.input)
		if want, have := testcase.want, testcase.input.Add(testcase.nodes...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.nodes, want, have)
		}
		if len(testcase.input) != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.nodes)
		}
	}
}

func BenchmarkNodeSetAdd(b *testing.B) {
	n := report.MakeNodeSet()
	for i := 0; i < 600; i++ {
		n = n.Add(
			report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
				"a": "1",
				"b": "2",
			}),
		)
	}

	node := report.MakeNode().WithID("401.5").WithLatests(map[string]string{
		"a": "1",
		"b": "2",
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult = n.Add(node)
	}
}

func TestNodeSetMerge(t *testing.T) {
	for _, testcase := range []struct {
		input report.NodeSet
		other report.NodeSet
		want  report.NodeSet
	}{
		{input: report.NodeSet(nil), other: report.NodeSet(nil), want: report.NodeSet(nil)},
		{input: report.MakeNodeSet(), other: report.MakeNodeSet(), want: report.MakeNodeSet()},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			other: report.MakeNodeSet(),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(),
			other: report.MakeNodeSet(report.MakeNode().WithID("a")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			other: report.MakeNodeSet(report.MakeNode().WithID("b")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("b")),
			other: report.MakeNodeSet(report.MakeNode().WithID("a")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a")),
			other: report.MakeNodeSet(report.MakeNode().WithID("a")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("c")),
			other: report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b"), report.MakeNode().WithID("c")),
		},
		{
			input: report.MakeNodeSet(report.MakeNode().WithID("b")),
			other: report.MakeNodeSet(report.MakeNode().WithID("a")),
			want:  report.MakeNodeSet(report.MakeNode().WithID("a"), report.MakeNode().WithID("b")),
		},
	} {
		originalLen := len(testcase.input)
		if want, have := testcase.want, testcase.input.Merge(testcase.other); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, want, have)
		}
		if len(testcase.input) != originalLen {
			t.Errorf("%v + %v: modified the original input!", testcase.input, testcase.other)
		}
	}
}

func BenchmarkNodeSetMerge(b *testing.B) {
	n, other := report.MakeNodeSet(), report.MakeNodeSet()
	for i := 0; i < 600; i++ {
		n = n.Add(
			report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
				"a": "1",
				"b": "2",
			}),
		)
	}

	for i := 400; i < 1000; i++ {
		other = other.Add(
			report.MakeNode().WithID(fmt.Sprint(i)).WithLatests(map[string]string{
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
