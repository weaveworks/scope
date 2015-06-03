package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
)

func TestLIFOReporter(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	ts := time.Now()
	now = func() time.Time { return ts }

	var (
		incoming = make(chan report.Report)
		maxAge   = time.Second
	)
	r := newLIFOReporter(incoming, maxAge)
	defer r.Stop()

	// By default, we want an empty report.
	want := report.MakeReport()
	if have := r.Report(); !reflect.DeepEqual(want, have) {
		t.Errorf("want\n\t%+v, have\n\t%+v", want, have)
	}

	// Seed the initial report.
	incoming <- report.Report{Endpoint: report.Topology{NodeMetadatas: report.NodeMetadatas{"a": report.NodeMetadata{"foo": "bar"}}}}
	want.Endpoint.NodeMetadatas["a"] = report.NodeMetadata{"foo": "bar"}
	if have := r.Report(); !reflect.DeepEqual(want, have) {
		t.Errorf("want\n\t%+v, have\n\t%+v", want, have)
	}

	// Add some time, and merge more reports.
	ts = ts.Add(maxAge / 2)
	incoming <- report.Report{Endpoint: report.Topology{NodeMetadatas: report.NodeMetadatas{"a": report.NodeMetadata{"1": "1"}}}}
	incoming <- report.Report{Endpoint: report.Topology{NodeMetadatas: report.NodeMetadatas{"a": report.NodeMetadata{"1": "1"}}}} // dupe!
	want.Endpoint.NodeMetadatas["a"] = report.NodeMetadata{"foo": "bar", "1": "1"}
	if have := r.Report(); !reflect.DeepEqual(want, have) {
		t.Errorf("want\n\t%+v, have\n\t%+v", want, have)
	}

	// Add enough time that the initial report should go away.
	ts = ts.Add(9 * (maxAge / 10))
	incoming <- report.MakeReport()
	want.Endpoint.NodeMetadatas["a"] = report.NodeMetadata{"1": "1"} // we should lose the initial data
	if have := r.Report(); !reflect.DeepEqual(want, have) {
		t.Errorf("want\n\t%+v, have\n\t%+v", want, have)
	}
}
