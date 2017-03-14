package report_test

import (
	"bytes"
	"compress/gzip"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestRoundtrip(t *testing.T) {
	var buf bytes.Buffer
	r1 := report.MakeReport()
	r1.WriteBinary(&buf, gzip.BestCompression)
	bytes := append([]byte{}, buf.Bytes()...) // copy the contents for later
	r2, err := report.MakeFromBinary(&buf)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
	r3, err := report.MakeFromBytes(bytes)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r3) {
		t.Errorf("%v != %v", r1, *r3)
	}
}

func TestRoundtripNoCompression(t *testing.T) {
	// Make sure that we can use our standard routines for decompressing
	// something with '0' level compression.
	var buf bytes.Buffer
	r1 := report.MakeReport()
	r1.WriteBinary(&buf, 0)
	r2, err := report.MakeFromBinary(&buf)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
}

func TestMoreCompressionMeansSmaller(t *testing.T) {
	// Make sure that 0 level compression actually does compress less.
	var buf1, buf2 bytes.Buffer
	r := report.MakeReport()
	r.WriteBinary(&buf1, gzip.BestCompression)
	r.WriteBinary(&buf2, 0)
	if buf1.Len() >= buf2.Len() {
		t.Errorf("Compression doesn't change size: %v >= %v", buf1.Len(), buf2.Len())
	}
}
