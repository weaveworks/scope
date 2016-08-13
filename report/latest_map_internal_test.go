package report

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

var (
	emptyString         = ""
	subSlabStringA      = strings.Repeat("a", slabSize-1)
	matchSlabStringA    = strings.Repeat("a", slabSize)
	overSlabStringA     = strings.Repeat("a", slabSize+1)
	subSlabStringB      = strings.Repeat("b", slabSize-1)
	matchSlabStringB    = strings.Repeat("b", slabSize)
	overSlabStringB     = strings.Repeat("b", slabSize+1)
	commonPrefixString1 = strings.Repeat("b", slabSize) + "aaa"
	commonPrefixString2 = strings.Repeat("b", slabSize) + "cc"
	testStrings         = []string{
		emptyString,
		subSlabStringA,
		matchSlabStringA, overSlabStringA,
		subSlabStringB, matchSlabStringB, overSlabStringB,
		commonPrefixString1, commonPrefixString2,
	}
)

func TestSlabStringString(t *testing.T) {
	for _, s := range testStrings {
		var ss slabString
		ss.fromString(s)
		if ss.String() != s {
			t.Errorf("%#q != %#q (expected)", ss, s)
		}
	}
}

func TestSlabStringBytes(t *testing.T) {
	for _, s := range testStrings {
		var ss slabString
		ss.fromBytes([]byte(s))
		if !bytes.Equal(ss.Bytes(), []byte(s)) {
			t.Errorf("%#q != %#q (expected)", ss, s)
		}
	}
}

func TestSlabStringEqual(t *testing.T) {
	for i := 0; i < len(testStrings); i++ {
		for j := i; j < len(testStrings); j++ {
			var a, b slabString
			a.fromString(testStrings[i])
			b.fromString(testStrings[j])
			have := a.Equal(b)
			want := bytes.Equal([]byte(testStrings[i]), []byte(testStrings[j]))
			if have != want {
				t.Errorf("Expected %#q == %#q to be %t", a, b, want)
			}

			// test commutativity
			if a.Equal(b) != b.Equal(a) {
				t.Errorf("Expected %#q == %#q to be commutative", a, b)
			}

		}
	}

}

func TestSlabStringCompare(t *testing.T) {
	for i := 0; i < len(testStrings); i++ {
		for j := i; j < len(testStrings); j++ {
			var a, b slabString
			a.fromString(testStrings[i])
			b.fromString(testStrings[j])
			have := a.compare(b)
			want := bytes.Compare([]byte(testStrings[i]), []byte(testStrings[j]))
			if have != want {
				t.Errorf("Expected Compare(%#q, %#q) to be %d (got %d)", a, b, want, have)
			}

			// test commutativity
			if a.compare(b) != -b.compare(a) {
				t.Errorf("Expected Compare(%#q, %#q) to be commutative", a, b)
			}
		}
	}

}

func TestSlabStringCompareBytes(t *testing.T) {
	for i := 0; i < len(testStrings); i++ {
		for j := i; j < len(testStrings); j++ {
			var a slabString
			a.fromString(testStrings[i])
			have := a.compareBytes([]byte(testStrings[j]))
			want := bytes.Compare([]byte(testStrings[i]), []byte(testStrings[j]))
			if have != want {
				t.Errorf("Expected CompareBytes(%#q, %#q) to be %d (got %d)", a, testStrings[j], want, have)
			}

			// test commutativity
			var b slabString
			b.fromString(testStrings[j])
			if a.compareBytes([]byte(testStrings[i])) != -b.compareBytes([]byte(testStrings[j])) {
				t.Errorf("Expected Compare(%#q, %#q) to be commutative", a, b)
			}
		}
	}

}

func TestLatestMapLookupIndex(t *testing.T) {
	m := EmptyLatestMap
	if i := m.lookupIndex([]byte("foo")); i != 0 {
		t.Errorf("Unexpected index: %d", i)
	}
	m = m.Set("foo", time.Now(), "fooValue")
	if i := m.lookupIndex([]byte("foo")); i != 0 {
		t.Errorf("Unexpected index: %d", i)
	}
	if i := m.lookupIndex([]byte("bar")); i != 0 {
		t.Errorf("Unexpected index: %d", i)
	}
	if i := m.lookupIndex([]byte("zaz")); i != 1 {
		t.Errorf("Unexpected index: %d", i)
	}

}

func TestLatestMapAdd(t *testing.T) {
	now := time.Now()
	have := EmptyLatestMap.
		Set("foo", now.Add(-1), "Baz").
		Set("foo", now, "Bar")
	if v, ok := have.Lookup("foo"); !ok || v != "Bar" {
		t.Errorf("v (%#q) != Bar", v)
	}
	if v, ok := have.Lookup("bar"); ok || v != "" {
		t.Errorf("v != nil")
	}
	have.ForEach(func(k, v string) {
		if k != "foo" || v != "Bar" {
			t.Errorf("v (%#q) != Bar", v)
		}
	})

	// Test commutativity
	have = EmptyLatestMap.
		Set("foo", now.Add(-1), "fooValue").
		Set("bar", now, "barValue")
	want := EmptyLatestMap.
		Set("bar", now, "barValue").
		Set("foo", now.Add(-1), "fooValue")

	if !reflect.DeepEqual(want, have) {
		t.Errorf(test.Diff(want, have))
	}

}

func TestLatestMapLookupEntry(t *testing.T) {
	haveT := time.Now()
	haveV := "Bar"
	have := EmptyLatestMap.Set("foo", haveT, haveV)
	if got, timestamp, ok := have.LookupEntry("foo"); !ok || got != haveV || !timestamp.Equal(haveT) {
		t.Errorf("got: %#v %v != expected %#v %v", got, timestamp, haveV, haveT)
	}
	if got, timestamp, ok := have.LookupEntry("not found"); ok {
		t.Errorf("found unexpected entry for %q: %#v %v", "not found", got, timestamp)
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

func BenchmarkLatestMapMerge(b *testing.B) {
	var (
		left  = EmptyLatestMap
		right = EmptyLatestMap
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
	want := EmptyLatestMap.
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
		have := EmptyLatestMap
		have.CodecDecodeSelf(decoder)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

}

func TestLatestMapEncodingNil(t *testing.T) {
	want := LatestMap{}

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
