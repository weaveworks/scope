package report

import (
	"bytes"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

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

func TestLatestMapLookupEntry(t *testing.T) {
	now := time.Now()
	entry := LatestEntry{Timestamp: now, Value: "Bar"}
	have := EmptyLatestMap.Set("foo", entry.Timestamp, entry.Value)
	if got, ok := have.LookupEntry("foo"); !ok || got != entry {
		t.Errorf("got: %#v != expected %#v", got, entry)
	}
	if got, ok := have.LookupEntry("not found"); ok {
		t.Errorf("found unexpected entry for %q: %#v", "not found", got)
	}
}

func TestLatestMapAddNil(t *testing.T) {
	now := time.Now()
	have := LatestMap{}.Set("foo", now, "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
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
	notequal := EmptyLatestMap.
		Set("foo", now, "Baz")
	if reflect.DeepEqual(want, notequal) {
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

func TestLatestMapDeleteNil(t *testing.T) {
	want := LatestMap{}
	have := LatestMap{}.Delete("foo")
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
		"nils": {
			a:    LatestMap{},
			b:    LatestMap{},
			want: LatestMap{},
		},
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

		for _, h := range []codec.Handle{
			codec.Handle(&codec.MsgpackHandle{}),
			codec.Handle(&codec.JsonHandle{}),
		} {
			buf := &bytes.Buffer{}
			encoder := codec.NewEncoder(buf, h)
			want.CodecEncodeSelf(encoder)
			decoder := codec.NewDecoder(buf, h)
			have := EmptyLatestMap
			have.CodecDecodeSelf(decoder)
			if !reflect.DeepEqual(want, have) {
				t.Error(test.Diff(want, have))
			}
		}
	}
}

func TestLatestMapEncodingNil(t *testing.T) {
	want := LatestMap{}

	{
		gobs, err := want.GobEncode()
		if err != nil {
			t.Fatal(err)
		}
		have := EmptyLatestMap
		have.GobDecode(gobs)
		if have.Map == nil {
			t.Error("Decoded LatestMap.psMap should not be nil")
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
			have := EmptyLatestMap
			have.CodecDecodeSelf(decoder)
			if !reflect.DeepEqual(want, have) {
				t.Error(test.Diff(want, have))
			}
		}
	}
}
