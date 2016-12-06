package report

import (
	"bytes"
	"testing"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/test"
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
	notequal := EmptyCounters.
		Add("foo", 4)
	if reflect.DeepEqual(want, notequal) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestCountersNil(t *testing.T) {
	want := Counters{}
	if want.Size() != 0 {
		t.Errorf("nil.Size != 0")
	}
	if v, ok := want.Lookup("foo"); ok || v != 0 {
		t.Errorf("nil.Lookup != false")
	}
	have := want.Add("foo", 1)
	if v, ok := have.Lookup("foo"); !ok || v != 1 {
		t.Errorf("nil.Add failed")
	}
	if have2 := want.Merge(have); !reflect.DeepEqual(have, have2) {
		t.Errorf(test.Diff(have, have2))
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

		for _, h := range []codec.Handle{
			codec.Handle(&codec.MsgpackHandle{}),
			codec.Handle(&codec.JsonHandle{}),
		} {
			buf := &bytes.Buffer{}
			encoder := codec.NewEncoder(buf, h)
			want.CodecEncodeSelf(encoder)
			decoder := codec.NewDecoder(buf, h)
			have := EmptyCounters
			have.CodecDecodeSelf(decoder)
			if !reflect.DeepEqual(want, have) {
				t.Error(test.Diff(want, have))
			}
		}
	}

}

func TestCountersString(t *testing.T) {
	{
		var c Counters
		have := c.String()
		want := `{}`
		if want != have {
			t.Errorf("Expected: %s, Got %s", want, have)
		}
	}

	{
		have := EmptyCounters.String()
		want := `{}`
		if want != have {
			t.Errorf("Expected: %s, Got %s", want, have)
		}
	}

	{
		have := EmptyCounters.
			Add("foo", 1).
			Add("bar", 2).String()

		want := `{bar: 2, foo: 1}`
		if want != have {
			t.Errorf("Expected: %s, Got %s", want, have)
		}
	}
}
