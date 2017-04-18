package appclient

import (
	"bytes"
	"compress/gzip"
	"github.com/weaveworks/scope/report"
)

// A ReportPublisher uses a buffer pool to serialise reports, which it
// then passes to a publisher
type ReportPublisher struct {
	publisher  Publisher
	noControls bool
}

// NewReportPublisher creates a new report publisher
func NewReportPublisher(publisher Publisher, noControls bool) *ReportPublisher {
	return &ReportPublisher{
		publisher:  publisher,
		noControls: noControls,
	}
}

// Publish serialises and compresses a report, then passes it to a publisher
func (p *ReportPublisher) Publish(r report.Report) error {
	if p.noControls {
		r.WalkTopologies(func(t *report.Topology) {
			t.Controls = report.Controls{}
		})
	}
	buf := &bytes.Buffer{}
	r.WriteBinary(buf, gzip.DefaultCompression)
	return p.publisher.Publish(buf)
}
