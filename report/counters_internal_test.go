package report

import (
	"testing"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestCountersAdd(t *testing.T) {
	have := EmptyCounters.
		Add("foo", 1).
		Add("foo", 2)
	if v, ok := have.Lookup("foo"); !ok || v != 3 {
		t.Errorf("foo != 3")
	}
	if v, ok := have.Lookup("bar"); ok || v != 0 {
		t.Errorf("bar != nil")
	}
}

func TestCountersDeepEquals(t *testing.T) {
	want := EmptyCounters.
		Add("foo", 3)
	have := EmptyCounters.
		Add("foo", 3)
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestCountersMerge(t *testing.T) {
	for name, c := range map[string]struct {
		a, b, want Counters
	}{
		"Empty a": {
			a: EmptyCounters,
			b: EmptyCounters.
				Add("foo", 1),
			want: EmptyCounters.
				Add("foo", 1),
		},
		"Empty b": {
			a: EmptyCounters.
				Add("foo", 1),
			b: EmptyCounters,
			want: EmptyCounters.
				Add("foo", 1),
		},
		"Disjoin a & b": {
			a: EmptyCounters.
				Add("foo", 1),
			b: EmptyCounters.
				Add("bar", 2),
			want: EmptyCounters.
				Add("foo", 1).
				Add("bar", 2),
		},
		"Overlapping a & b": {
			a: EmptyCounters.
				Add("foo", 1),
			b: EmptyCounters.
				Add("foo", 2),
			want: EmptyCounters.
				Add("foo", 3),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func TestCountersEncoding(t *testing.T) {
	want := EmptyCounters.
		Add("foo", 1).
		Add("bar", 2)

	{
		gobs, err := want.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyCounters
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
		have := EmptyCounters
		have.UnmarshalJSON(json)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}
