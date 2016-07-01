package appclient

import (
	"bytes"
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

func removeControls(r report.Report) report.Report {
	r.Endpoint.Controls = report.Controls{}
	r.Process.Controls = report.Controls{}
	r.Container.Controls = report.Controls{}
	r.ContainerImage.Controls = report.Controls{}
	r.Pod.Controls = report.Controls{}
	r.Service.Controls = report.Controls{}
	r.Deployment.Controls = report.Controls{}
	r.ReplicaSet.Controls = report.Controls{}
	r.Host.Controls = report.Controls{}
	r.Overlay.Controls = report.Controls{}
	return r
}

// Publish serialises and compresses a report, then passes it to a publisher
func (p *ReportPublisher) Publish(r report.Report) error {
	if p.noControls {
		r = removeControls(r)
	}
	buf := &bytes.Buffer{}
	r.WriteBinary(buf)
	return p.publisher.Publish(buf)
}
