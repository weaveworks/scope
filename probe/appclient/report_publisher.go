package appclient

import (
	"github.com/weaveworks/scope/report"
)

// A ReportPublisher uses a buffer pool to serialise reports, which it
// then passes to a publisher
type ReportPublisher struct {
	publisher Publisher
}

// NewReportPublisher creates a new report publisher
func NewReportPublisher(publisher Publisher) *ReportPublisher {
	return &ReportPublisher{
		publisher: publisher,
	}
}

// Publish serialises and compresses a report, then passes it to a publisher
func (p *ReportPublisher) Publish(r report.Report) error {
	return p.publisher.Publish(r)
}
