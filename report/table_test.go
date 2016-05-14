package report_test

import (
	"fmt"
	"reflect"
	"testing"

	"$GITHUB_URI/report"
	"$GITHUB_URI/test"
)

func TestTables(t *testing.T) {
	want := map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}
	nmd := report.MakeNode("foo1")

	nmd = nmd.AddTable("foo_", want)
	have, truncationCount := nmd.ExtractTable("foo_")

	if truncationCount != 0 {
		t.Error("Table shouldn't had been truncated")
	}

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestTruncation(t *testing.T) {
	wantTruncationCount := 1
	want := map[string]string{}
	for i := 0; i < report.MaxTableRows+wantTruncationCount; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		want[key] = value
	}

	nmd := report.MakeNode("foo1")

	nmd = nmd.AddTable("foo_", want)
	_, truncationCount := nmd.ExtractTable("foo_")

	if truncationCount != wantTruncationCount {
		t.Error(
			"Table should had been truncated by",
			wantTruncationCount,
			"and not",
			truncationCount,
		)
	}
}
