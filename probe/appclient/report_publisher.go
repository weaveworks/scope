package appclient

import (
	"bytes"
	"compress/gzip"

	"github.com/weaveworks/scope/report"
)

// ReportPublisher publishes reports, probably to a remote collector.
type ReportPublisher interface {
	Publish(r report.Report) error
}

type reportPublisher struct {
	publisher  ReportPublisher
	noControls bool
}

// NewReportPublisher creates a new report publisher
func NewReportPublisher(publisher ReportPublisher, noControls bool) ReportPublisher {
	return &reportPublisher{
		publisher:  publisher,
		noControls: noControls,
	}
}

func serializeReport(r report.Report) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	err := r.WriteBinary(buf, gzip.DefaultCompression)
	return buf, err
}

// Publish sanitises a report, then passes it to a publisher
func (p *reportPublisher) Publish(r report.Report) error {
	if p.noControls {
		r.WalkTopologies(func(t *report.Topology) {
			t.Controls = report.Controls{}
		})
	}
	return p.publisher.Publish(r)
}
