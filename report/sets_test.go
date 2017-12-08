package report_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestSetsAdd(t *testing.T) {
	for _, testcase := range []struct {
		a    report.Sets
		want map[string][]string
	}{
		{
			report.MakeSets().Add("a", report.MakeStringSet("b")),
			map[string][]string{"a": {"b"}},
		},
		{
			report.MakeSets().Add("a", report.MakeStringSet("b")).Add("a", report.MakeStringSet("c")),
			map[string][]string{"a": {"b", "c"}},
		},
		{
			report.MakeSets().Add("a", report.MakeStringSet("b", "c")).Add("a", report.MakeStringSet("c")),
			map[string][]string{"a": {"b", "c"}},
		},
		{
			report.MakeSets().Add("a", report.MakeStringSet("c")).Add("a", report.MakeStringSet("b", "c")),
			map[string][]string{"a": {"b", "c"}},
		},
		{
			report.MakeSets().Add("a", report.MakeStringSet("1")).Add("b", report.MakeStringSet("2")).
				Add("c", report.MakeStringSet("3")).Add("b", report.MakeStringSet("3")),
			map[string][]string{"a": {"1"}, "b": {"2", "3"}, "c": {"3"}},
		},
	} {
		check(t, "Add", testcase.a, testcase.want)
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
		check(t, fmt.Sprintf("%+v.Merge(%+v)", testcase.a, testcase.b), testcase.a.Merge(testcase.b), testcase.want)
		check(t, fmt.Sprintf("%+v.Merge(%+v)", testcase.b, testcase.a), testcase.b.Merge(testcase.a), testcase.want)
	}
}

func check(t *testing.T, desc string, haveSets report.Sets, want map[string][]string) {
	if haveSets.Size() != len(want) {
		t.Errorf("%s: different lengths: want %+v, have %+v", desc, want, haveSets)
	}
	for k, v := range want {
		have, _ := haveSets.Lookup(k)
		if !reflect.DeepEqual([]string(have), v) {
			t.Errorf("%s: want %+v, have %+v", desc, want, haveSets)
		}
	}
}
