package report_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestRoundtrip(t *testing.T) {
	var buf bytes.Buffer
	r1 := report.MakeReport()
	r1.WriteBinary(&buf)
	r2, err := report.MakeFromBinary(&buf)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
}
