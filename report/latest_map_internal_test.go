package report

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestLatestMapAdd(t *testing.T) {
	now := time.Now()
	have := EmptyStringLatestMap.
		Set("foo", now.Add(-1), "Baz").
		Set("foo", now, "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
	if v, ok := have.Lookup("bar"); ok || v != "" {
		t.Errorf("v != nil")
	}
	have.ForEach(func(k string, _ time.Time, v string) {
		if k != "foo" || v != "Bar" {
			t.Errorf("v != Bar")
		}
	})
}

func TestLatestMapLookupEntry(t *testing.T) {
	now := time.Now()
	type LatestEntry struct {
		Timestamp time.Time
		Value     interface{}
	}
	entry := LatestEntry{Timestamp: now, Value: "Bar"}
	have := EmptyStringLatestMap.Set("foo", entry.Timestamp, entry.Value.(string))
	if got, timestamp, ok := have.LookupEntry("foo"); !ok || got != entry.Value || !timestamp.Equal(entry.Timestamp) {
		t.Errorf("got: %#v %v != expected %#v", got, timestamp, entry)
	}
	if got, timestamp, ok := have.LookupEntry("not found"); ok {
		t.Errorf("found unexpected entry for %q: %#v %v", "not found", got, timestamp)
	}
}

func TestLatestMapAddNil(t *testing.T) {
	now := time.Now()
	have := StringLatestMap{}.Set("foo", now, "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
}

func TestLatestMapDeepEquals(t *testing.T) {
	now := time.Now()
	want := EmptyStringLatestMap.
		Set("foo", now, "Bar")
	have := EmptyStringLatestMap.
		Set("foo", now, "Bar")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
	notequal := EmptyStringLatestMap.
		Set("foo", now, "Baz")
	if reflect.DeepEqual(want, notequal) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestLatestMapDelete(t *testing.T) {
	now := time.Now()
	want := EmptyStringLatestMap
	have := EmptyStringLatestMap.
		Set("foo", now, "Baz").
		Delete("foo")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
}

func TestLatestMapDeleteNil(t *testing.T) {
	want := StringLatestMap{}
	have := StringLatestMap{}.Delete("foo")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
}

func nilStringLatestMap() StringLatestMap {
	m := EmptyStringLatestMap
	m.Map = nil
	return m
}

func TestLatestMapMerge(t *testing.T) {
	now := time.Now()
	then := now.Add(-1)

	for name, c := range map[string]struct {
		a, b, want StringLatestMap
	}{
		"nils": {
			a:    nilStringLatestMap(),
			b:    nilStringLatestMap(),
			want: nilStringLatestMap(),
		},
		"Empty a": {
			a: EmptyStringLatestMap,
			b: EmptyStringLatestMap.
				Set("foo", now, "bar"),
			want: EmptyStringLatestMap.
				Set("foo", now, "bar"),
		},
		"Empty b": {
			a: EmptyStringLatestMap.
				Set("foo", now, "bar"),
			b: EmptyStringLatestMap,
			want: EmptyStringLatestMap.
				Set("foo", now, "bar"),
		},
		"Disjoint a & b": {
			a: EmptyStringLatestMap.
				Set("foo", now, "bar"),
			b: EmptyStringLatestMap.
				Set("baz", now, "bop"),
			want: EmptyStringLatestMap.
				Set("foo", now, "bar").
				Set("baz", now, "bop"),
		},
		"Common a & b": {
			a: EmptyStringLatestMap.
				Set("foo", now, "bar"),
			b: EmptyStringLatestMap.
				Set("foo", then, "baz"),
			want: EmptyStringLatestMap.
				Set("foo", now, "bar"),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func BenchmarkLatestMapMerge(b *testing.B) {
	var (
		left  = EmptyStringLatestMap
		right = EmptyStringLatestMap
		now   = time.Now()
	)

	// two large maps with some overlap
	for i := 0; i < 1000; i++ {
		left = left.Set(fmt.Sprint(i), now, "1")
	}
	for i := 700; i < 1700; i++ {
		right = right.Set(fmt.Sprint(i), now.Add(1*time.Minute), "1")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		left.Merge(right)
	}
}

func TestLatestMapEncoding(t *testing.T) {
	now := time.Now()
	want := EmptyStringLatestMap.
		Set("foo", now, "bar").
		Set("bar", now, "baz")

	for _, h := range []codec.Handle{
		codec.Handle(&codec.MsgpackHandle{}),
		codec.Handle(&codec.JsonHandle{}),
	} {
		buf := &bytes.Buffer{}
		encoder := codec.NewEncoder(buf, h)
		want.CodecEncodeSelf(encoder)
		decoder := codec.NewDecoder(buf, h)
		have := EmptyStringLatestMap
		have.CodecDecodeSelf(decoder)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

}

func TestLatestMapEncodingNil(t *testing.T) {
	want := nilStringLatestMap()

	for _, h := range []codec.Handle{
		codec.Handle(&codec.MsgpackHandle{}),
		codec.Handle(&codec.JsonHandle{}),
	} {
		buf := &bytes.Buffer{}
		encoder := codec.NewEncoder(buf, h)
		want.CodecEncodeSelf(encoder)
		decoder := codec.NewDecoder(buf, h)
		have := EmptyStringLatestMap
		have.CodecDecodeSelf(decoder)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

}

func TestLatestMapMergeEqualDecoderTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("Merging two maps with the same decoders should not panic")
		}
	}()
	m1 := MakeStringLatestMap().Set("a", time.Now(), "bar")
	m2 := MakeStringLatestMap().Set("b", time.Now(), "foo")
	m1.Merge(m2)
}
