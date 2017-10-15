package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestSetsDelete(t *testing.T) {
	aSet := report.MakeSets().Add("a", report.MakeStringSet("a"))
	for _, testcase := range []struct {
		input report.Sets
		key   string
		want  report.Sets
	}{
		{key: "", input: report.Sets{}, want: report.Sets{}},
		{key: "", input: report.MakeSets(), want: report.MakeSets()},
		{key: "", input: aSet, want: aSet},
		{key: "a", input: report.MakeSets(), want: report.MakeSets()},
		{key: "a", input: aSet, want: report.MakeSets()},
		{key: "b", input: aSet, want: aSet},
		{key: "b", input: aSet.Add("b", report.MakeStringSet("b")), want: aSet},
		{
			input: aSet.Add("b", report.MakeStringSet("b")),
			key:   "a",
			want:  report.MakeSets().Add("b", report.MakeStringSet("b")),
		},
	} {
		originalLen := testcase.input.Size()
		if want, have := testcase.want, testcase.input.Delete(testcase.key); !reflect.DeepEqual(want, have) {
			t.Errorf("%v - %v: want %v, have %v", testcase.input, testcase.key, want, have)
		}
		if testcase.input.Size() != originalLen {
			t.Errorf("%v - %v: modified the original input!", testcase.input, testcase.key)
		}
	}
}

func TestSetsMerge(t *testing.T) {
	for _, testcase := range []struct {
		a, b report.Sets
		want map[string][]string
	}{
		{report.MakeSets(), report.MakeSets(), map[string][]string{}},
		{
			report.MakeSets(),
			report.MakeSets().Add("a", report.MakeStringSet("b")),
			map[string][]string{"a": {"b"}},
		},
		{
			report.MakeSets(),
			report.MakeSets().Add("a", report.MakeStringSet("b", "c")),
			map[string][]string{"a": {"b", "c"}},
		},
		{
			report.MakeSets().Add("a", report.MakeStringSet("1")).Add("b", report.MakeStringSet("2")),
			report.MakeSets().Add("c", report.MakeStringSet("3")).Add("b", report.MakeStringSet("3")),
			map[string][]string{"a": {"1"}, "b": {"2", "3"}, "c": {"3"}},
		},
	} {
		haveSets := testcase.a.Merge(testcase.b)
		have := map[string][]string{}
		keys := haveSets.Keys()
		for _, k := range keys {
			have[k], _ = haveSets.Lookup(k)
		}

		if !reflect.DeepEqual(testcase.want, have) {
			t.Errorf("%+v.Merge(%+v): want %+v, have %+v", testcase.a, testcase.b, testcase.want, have)
		}
	}
}
