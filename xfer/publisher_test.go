package xfer_test

import (
	"io/ioutil"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func TestGzipGobEncoder(t *testing.T) {
	if err := xfer.GzipGobEncoder(ioutil.Discard, report.MakeReport()); err != nil {
		t.Error(err)
	}
}

func TestJSONEncoder(t *testing.T) {
	if err := xfer.JSONEncoder(ioutil.Discard, report.MakeReport()); err != nil {
		t.Error(err)
	}
}

func TestSendingPublisher(t *testing.T) {
	var (
		enc       = xfer.JSONEncoder
		sender    = &mockSender{}
		publisher = xfer.NewSendingPublisher(enc, sender)
		rpt       = report.MakeReport()
	)
	rpt.Endpoint.AddNode("foo", report.MakeNodeWith(map[string]string{"bar": "baz"}))
	if err := publisher.Publish(rpt); err != nil {
		t.Error(err)
	}
	if sender.buf.Len() <= 0 {
		t.Errorf("mockSender hasn't received anything")
	}
}
