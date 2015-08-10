package xfer_test

import (
	"reflect"
	"time"

	"github.com/weaveworks/scope/test"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"

	"testing"
)

func TestCollector(t *testing.T) {
	window := time.Millisecond
	c := xfer.NewCollector(window)

	r1 := report.MakeReport()
	r1.Endpoint.NodeMetadatas["foo"] = report.MakeNodeMetadata()

	r2 := report.MakeReport()
	r2.Endpoint.NodeMetadatas["bar"] = report.MakeNodeMetadata()

	if want, have := report.MakeReport(), c.Report(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	c.Add(r1)
	if want, have := r1, c.Report(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	c.Add(r2)
	merged := report.MakeReport()
	merged.Merge(r1)
	merged.Merge(r2)
	if want, have := merged, c.Report(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
