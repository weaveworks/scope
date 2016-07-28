package report

import (
	"bytes"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestConstMapEncoding(t *testing.T) {
	now := time.Now()
	want := EmptyConstMap.
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
		have := EmptyConstMap
		have.CodecDecodeSelf(decoder)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

}

func TestConstMapEncodingNil(t *testing.T) {
	want := ConstMap{}

	jsonHandle := codec.Handle(&codec.JsonHandle{})
	for _, h := range []codec.Handle{
		jsonHandle,
		codec.Handle(&codec.MsgpackHandle{}),
	} {
		buf := &bytes.Buffer{}
		encoder := codec.NewEncoder(buf, h)
		want.CodecEncodeSelf(encoder)

		if h == jsonHandle && buf.String() != "{}" {
			t.Error("Non-empty map when encoding empty ConstMap:", buf.String())
		}

		decoder := codec.NewDecoder(buf, h)
		have := EmptyConstMap
		have.CodecDecodeSelf(decoder)
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}

}
