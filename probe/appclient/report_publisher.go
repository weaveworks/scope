package appclient

import (
	"bytes"
	"compress/gzip"
	"github.com/ugorji/go/codec"

	"$GITHUB_URI/report"
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
	buf := &bytes.Buffer{}
	gzwriter := gzip.NewWriter(buf)
	if err := codec.NewEncoder(gzwriter, &codec.MsgpackHandle{}).Encode(r); err != nil {
		return err
	}
	gzwriter.Close() // otherwise the content won't get flushed to the output stream

	return p.publisher.Publish(buf)
}
