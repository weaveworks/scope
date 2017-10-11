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
	have := MakeStringLatestMap().
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
	have := MakeStringLatestMap().Set("foo", entry.Timestamp, entry.Value.(string))
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
	want := MakeStringLatestMap().
		Set("foo", now, "Bar")
	have := MakeStringLatestMap().
		Set("foo", now, "Bar")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
	notequal := MakeStringLatestMap().
		Set("foo", now, "Baz")
	if reflect.DeepEqual(want, notequal) {
		t.Errorf(test.Diff(want, have))
	}
}

func nilStringLatestMap() StringLatestMap {
	m := MakeStringLatestMap()
	m.entries = nil
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
			a: MakeStringLatestMap(),
			b: MakeStringLatestMap().
				Set("foo", now, "bar"),
			want: MakeStringLatestMap().
				Set("foo", now, "bar"),
		},
		"Empty b": {
			a: MakeStringLatestMap().
				Set("foo", now, "bar"),
			b: MakeStringLatestMap(),
			want: MakeStringLatestMap().
				Set("foo", now, "bar"),
		},
		"Disjoint a & b": {
			a: MakeStringLatestMap().
				Set("foo", now, "bar"),
			b: MakeStringLatestMap().
				Set("baz", now, "bop"),
			want: MakeStringLatestMap().
				Set("foo", now, "bar").
				Set("baz", now, "bop"),
		},
		"Common a & b": {
			a: MakeStringLatestMap().
				Set("foo", now, "bar"),
			b: MakeStringLatestMap().
				Set("foo", then, "baz"),
			want: MakeStringLatestMap().
				Set("foo", now, "bar"),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func makeBenchmarkMap(start, finish int, timestamp time.Time) StringLatestMap {
	ret := MakeStringLatestMap()
	for i := start; i < finish; i++ {
		ret = ret.Set(fmt.Sprint(i), timestamp, "1")
	}
	return ret
}

func BenchmarkLatestMapMerge(b *testing.B) {
	// two large maps with some overlap
	left := makeBenchmarkMap(0, 1000, time.Now())
	right := makeBenchmarkMap(700, 1700, time.Now().Add(1*time.Minute))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		left.Merge(right)
	}
}

func BenchmarkLatestMapEncode(b *testing.B) {
	map1 := makeBenchmarkMap(0, 1000, time.Now())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(&map1)
	}
}

func BenchmarkLatestMapDecode(b *testing.B) {
	map1 := makeBenchmarkMap(0, 1000, time.Now())
	buf := &bytes.Buffer{}
	codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(&map1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var map1 StringLatestMap
		codec.NewDecoderBytes(buf.Bytes(), &codec.MsgpackHandle{}).Decode(&map1)
	}
}

func TestLatestMapEncoding(t *testing.T) {
	now := time.Now()
	want := MakeStringLatestMap().
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
		have := MakeStringLatestMap()
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
		have := MakeStringLatestMap()
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
