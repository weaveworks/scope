package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestMakeStringSet(t *testing.T) {
	for _, testcase := range []struct {
		input []string
		want  report.StringSet
	}{
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
		{input: report.MakeStringSet(), other: report.MakeStringSet(), want: report.MakeStringSet()},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet(), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet(), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("b"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a", "c"), other: report.MakeStringSet("a", "b"), want: report.MakeStringSet("a", "b", "c")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
	} {
		if want, have := testcase.want, testcase.input.Merge(testcase.other); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, want, have)
		}
	}

}
