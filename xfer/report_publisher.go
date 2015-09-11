package xfer

import (
	"compress/gzip"
	"encoding/gob"
	"sync"

	"github.com/weaveworks/scope/report"
)

// A ReportPublisher uses a buffer pool to serialise reports, which it
// then passes to a publisher
type ReportPublisher struct {
	buffers   *sync.Pool
	publisher Publisher
}

// NewReportPublisher creates a new report publisher
func NewReportPublisher(publisher Publisher) *ReportPublisher {
	return &ReportPublisher{
		buffers:   NewBufferPool(),
		publisher: publisher,
	}
}

// Publish serialises and compresses a report, then passes it to a publisher
func (p *ReportPublisher) Publish(r report.Report) error {
	buf := p.buffers.Get().(*Buffer)
	gzwriter := gzip.NewWriter(buf)
	if err := gob.NewEncoder(gzwriter).Encode(r); err != nil {
		buf.Reset()
		p.buffers.Put(buf)
		return err
	}
	gzwriter.Close() // otherwise the content won't get flushed to the output stream

	return p.publisher.Publish(buf)
}
