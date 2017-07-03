package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

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
