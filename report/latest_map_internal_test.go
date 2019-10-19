package report

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestLatestMapAdd(t *testing.T) {
	have := MakeStringLatestMap().
		Set("foo", "Baz").
		Set("foo", "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
	if v, ok := have.Lookup("bar"); ok || v != "" {
		t.Errorf("v != nil")
	}
	have.ForEach(func(k string, v string) {
		if k != "foo" || v != "Bar" {
			t.Errorf("v != Bar")
		}
	})
}

func TestLatestMapLookupEntry(t *testing.T) {
	value := "Bar"
	have := MakeStringLatestMap().Set("foo", value)
	if got, ok := have.Lookup("foo"); !ok || got != value {
		t.Errorf("got: %#v != expected %#v", got, value)
	}
	if got, ok := have.Lookup("not found"); ok {
		t.Errorf("found unexpected entry for %q: %#v", "not found", got)
	}
}

func TestLatestMapAddNil(t *testing.T) {
	have := StringLatestMap{}.Set("foo", "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v != Bar")
	}
}

func TestLatestMapDeepEquals(t *testing.T) {
	want := MakeStringLatestMap().
		Set("foo", "Bar")
	have := MakeStringLatestMap().
		Set("foo", "Bar")
	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}
	notequal := MakeStringLatestMap().
		Set("foo", "Baz")
	if reflect.DeepEqual(want, notequal) {
		t.Errorf(test.Diff(want, have))
	}
}

func nilStringLatestMap() StringLatestMap {
	return nil
}

func TestLatestMapMerge(t *testing.T) {
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
				Set("foo", "bar"),
			want: MakeStringLatestMap().
				Set("foo", "bar"),
		},
		"Identical a & b": {
			a: MakeStringLatestMap().
				Set("foo", "bar"),
			b: MakeStringLatestMap().
				Set("foo", "bar"),
			want: MakeStringLatestMap().
				Set("foo", "bar"),
		},
		"Disjoint a & b": {
			a: MakeStringLatestMap().
				Set("foo", "bar"),
			b: MakeStringLatestMap().
				Set("baz", "bop"),
			want: MakeStringLatestMap().
				Set("foo", "bar").
				Set("baz", "bop"),
		},
		"Common a & b": { // b overrides a where there are keys in common
			a: MakeStringLatestMap().
				Set("foo", "baz"),
			b: MakeStringLatestMap().
				Set("foo", "bar"),
			want: MakeStringLatestMap().
				Set("foo", "bar"),
		},
		"Longer": { // b overrides a where there are keys in common
			a: MakeStringLatestMap().
				Set("PID", "0").
				Set("Name", "curl"),
			b: MakeStringLatestMap().
				Set("PID", "23128").
				Set("Name", "curl").
				Set("Domain", "node-a.local"),
			want: MakeStringLatestMap().
				Set("PID", "23128").
				Set("Name", "curl").
				Set("Domain", "node-a.local"),
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s:\n%s", name, test.Diff(c.want, have))
		}
	}
}

func makeBenchmarkMap(start, finish int) StringLatestMap {
	ret := MakeStringLatestMap()
	for i := start; i < finish; i++ {
		ret = ret.Set(fmt.Sprint(i), "1")
	}
	return ret
}

func BenchmarkLatestMapMerge(b *testing.B) {
	// two large maps with some overlap
	left := makeBenchmarkMap(0, 1000)
	right := makeBenchmarkMap(700, 1700)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		left.Merge(right)
	}
}

func BenchmarkLatestMapEncode(b *testing.B) {
	map1 := makeBenchmarkMap(0, 1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(&map1)
	}
}

func BenchmarkLatestMapDecode(b *testing.B) {
	map1 := makeBenchmarkMap(0, 1000)
	buf := &bytes.Buffer{}
	codec.NewEncoder(buf, &codec.MsgpackHandle{}).Encode(&map1)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var map1 StringLatestMap
		codec.NewDecoderBytes(buf.Bytes(), &codec.MsgpackHandle{}).Decode(&map1)
	}
}

func TestLatestMapDecoding(t *testing.T) {
	want := MakeStringLatestMap().
		Set("foo", "bar").
		Set("bar", "baz").
		Set("emptyval", "")
	// The following string is carefully constructed to have 'emptyval' not in alphabetical order
	data := `
{
  "bar": "baz",
  "foo": "bar",
  "emptyval": ""
}`
	h := &codec.JsonHandle{}
	decoder := codec.NewDecoder(bytes.NewBufferString(data), h)
	have := MakeStringLatestMap()
	have.CodecDecodeSelf(decoder)
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestLatestMapEncoding(t *testing.T) {
	want := MakeStringLatestMap().
		Set("foo", "bar").
		Set("bar", "baz")

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
	m1 := MakeStringLatestMap().Set("a", "bar")
	m2 := MakeStringLatestMap().Set("b", "foo")
	m1.Merge(m2)
}
