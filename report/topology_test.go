package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestMakeStringSet(t *testing.T) {
	for _, testcase := range []struct {
		input []string
		want  report.StringSet
	}{
		{input: nil, want: nil},
		{input: []string{}, want: report.MakeStringSet()},
		{input: []string{"a"}, want: report.MakeStringSet("a")},
		{input: []string{"a", "a"}, want: report.MakeStringSet("a")},
		{input: []string{"b", "c", "a"}, want: report.MakeStringSet("a", "b", "c")},
	} {
		if want, have := testcase.want, report.MakeStringSet(testcase.input...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v: want %v, have %v", testcase.input, want, have)
		}
	}
}

func TestStringSetAdd(t *testing.T) {
	for _, testcase := range []struct {
		input report.StringSet
		strs  []string
		want  report.StringSet
	}{
		{input: report.StringSet(nil), strs: []string{}, want: report.StringSet(nil)},
		{input: report.MakeStringSet(), strs: []string{}, want: report.MakeStringSet()},
		{input: report.MakeStringSet("a"), strs: []string{}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet(), strs: []string{"a"}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a"), strs: []string{"a"}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("b"), strs: []string{"a", "b"}, want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("a"), strs: []string{"c", "b"}, want: report.MakeStringSet("a", "b", "c")},
		{input: report.MakeStringSet("a", "c"), strs: []string{"b", "b", "b"}, want: report.MakeStringSet("a", "b", "c")},
	} {
		if want, have := testcase.want, testcase.input.Add(testcase.strs...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.strs, want, have)
		}
	}
}

func TestStringSetMerge(t *testing.T) {
	for _, testcase := range []struct {
		input report.StringSet
		other report.StringSet
		want  report.StringSet
	}{
		{input: report.StringSet(nil), other: report.StringSet(nil), want: report.StringSet(nil)},
		{input: report.MakeStringSet(), other: report.MakeStringSet(), want: report.MakeStringSet()},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet(), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet(), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("b"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a", "c"), other: report.MakeStringSet("a", "b"), want: report.MakeStringSet("a", "b", "c")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
	} {
		if have, _ := testcase.input.Merge(testcase.other); !reflect.DeepEqual(testcase.want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, testcase.want, have)
		}
	}
}

func TestNodeOrdering(t *testing.T) {
	ids := [][2]string{{}, {"a", "0"}, {"a", "1"}, {"b", "0"}, {"b", "1"}, {"c", "3"}}
	nodes := []report.Node{}
	for _, id := range ids {
		nodes = append(nodes, report.MakeNode(id[1]).WithTopology(id[0]))
	}

	for i, node := range nodes {
		if !node.Equal(node) {
			t.Errorf("Expected %q %q == %q %q, but was not", node.Topology, node.ID, node.Topology, node.ID)
		}
		if i > 0 {
			if !node.After(nodes[i-1]) {
				t.Errorf("Expected %q %q > %q %q, but was not", node.Topology, node.ID, nodes[i-1].Topology, nodes[i-1].ID)
			}
			if !nodes[i-1].Before(node) {
				t.Errorf("Expected %q %q < %q %q, but was not", nodes[i-1].Topology, nodes[i-1].ID, node.Topology, node.ID)
			}
		}
	}
}
