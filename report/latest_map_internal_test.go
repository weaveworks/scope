package report

import (
	"testing"
	"time"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestLatestMapAdd(t *testing.T) {
	now := time.Now()
	have := EmptyLatestMap.
		Set("foo", now.Add(-1), "Baz").
		Set("foo", now, "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
	if v, ok := have.Lookup("bar"); ok || v != "" {
		t.Errorf("v != nil")
	}
	have.ForEach(func(k, v string) {
		if k != "foo" || v != "Bar" {
			t.Errorf("v != Bar")
		}
	})
}

func TestLatestMapDeepEquals(t *testing.T) {
	now := time.Now()
	want := EmptyLatestMap.
		Set("foo", now, "Bar")
	have := EmptyLatestMap.
		Set("foo", now, "Bar")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestLatestMapDelete(t *testing.T) {
	now := time.Now()
	want := EmptyLatestMap
	have := EmptyLatestMap.
		Set("foo", now, "Baz").
		Delete("foo")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestLatestMapMerge(t *testing.T) {
	now := time.Now()
	then := now.Add(-1)

	for name, c := range map[string]struct {
		a, b, want LatestMap
	}{
		"Empty a": {
			a: EmptyLatestMap,
			b: EmptyLatestMap.
				Set("foo", now, "bar"),
			want: EmptyLatestMap.
				Set("foo", now, "bar"),
		},
		"Empty b": {
			a: EmptyLatestMap.
				Set("foo", now, "bar"),
			b: EmptyLatestMap,
			want: EmptyLatestMap.
				Set("foo", now, "bar"),
		},
		"Disjoint a & b": {
			a: EmptyLatestMap.
				Set("foo", now, "bar"),
			b: EmptyLatestMap.
				Set("baz", now, "bop"),
			want: EmptyLatestMap.
				Set("foo", now, "bar").
				Set("baz", now, "bop"),
		},
		"Common a & b": {
			a: EmptyLatestMap.
				Set("foo", now, "bar"),
			b: EmptyLatestMap.
				Set("foo", then, "baz"),
			want: EmptyLatestMap.
				Set("foo", now, "bar"),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func TestLatestMapEncoding(t *testing.T) {
	now := time.Now()
	want := EmptyLatestMap.
		Set("foo", now, "bar").
		Set("bar", now, "baz")

	{
		gobs, err := want.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyLatestMap
		have.GobDecode(gobs)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

	{
		json, err := want.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyLatestMap
		have.UnmarshalJSON(json)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}
