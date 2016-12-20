package report_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
)

func TestPrefixTables(t *testing.T) {
	want := map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}
	nmd := report.MakeNode("foo1")

	nmd = nmd.AddPrefixPropertyList("foo_", want)
	have, truncationCount := nmd.ExtractPropertyList(report.PropertyListTemplate{Prefix: "foo_"})

	if truncationCount != 0 {
		t.Error("Table shouldn't had been truncated")
	}

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestFixedTables(t *testing.T) {
	want := map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}
	nmd := report.MakeNodeWith("foo1", map[string]string{
		"foo1key": "bar1",
		"foo2key": "bar2",
	})

	template := report.PropertyListTemplate{FixedProperties: map[string]string{
		"foo1key": "foo1",
		"foo2key": "foo2",
	},
	}

	have, _ := nmd.ExtractPropertyList(template)

	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestTruncation(t *testing.T) {
	wantTruncationCount := 1
	want := map[string]string{}
	for i := 0; i < report.MaxPropertyListSize+wantTruncationCount; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		want[key] = value
	}

	nmd := report.MakeNode("foo1")

	nmd = nmd.AddPrefixPropertyList("foo_", want)
	_, truncationCount := nmd.ExtractPropertyList(report.PropertyListTemplate{Prefix: "foo_"})

	if truncationCount != wantTruncationCount {
		t.Error(
			"Table should had been truncated by",
			wantTruncationCount,
			"and not",
			truncationCount,
		)
	}
}
