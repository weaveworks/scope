package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestStringSetContains(t *testing.T) {
	for _, testcase := range []struct {
		contents []string
		target   string
		want     bool
	}{
		{nil, "foo", false},
		{[]string{}, "foo", false},
		{[]string{"a"}, "foo", false},
		{[]string{"a", "foo"}, "foo", true},
		{[]string{"foo", "b"}, "foo", true},
	} {
		have := report.MakeStringSet(testcase.contents...).Contains(testcase.target)
		if testcase.want != have {
			t.Errorf("%+v.Contains(%q): want %v, have %v", testcase.contents, testcase.target, testcase.want, have)
		}
	}
}
