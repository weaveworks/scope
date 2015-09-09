package xfer

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"sync"
	"sync/atomic"

	"github.com/weaveworks/scope/report"
)

// A Buffer is a reference counted bytes.Buffer, which belongs
// to a sync.Pool
type Buffer struct {
	bytes.Buffer
	pool *sync.Pool
	refs int32
}

// NewBuffer creates a new buffer
func NewBuffer(pool *sync.Pool) *Buffer {
	return &Buffer{
		pool: pool,
		refs: 0,
	}
}

// Get increases the reference count.  It is safe for concurrent calls.
func (b *Buffer) Get() {
	atomic.AddInt32(&b.refs, 1)
}

// Put decreases the reference count, and when it hits zero, puts the
// buffer back in the pool.
func (b *Buffer) Put() {
	if atomic.AddInt32(&b.refs, -1) == 0 {
		b.Reset()
		b.pool.Put(b)
	}
}

// NewBufferPool creates a new buffer pool.
func NewBufferPool() *sync.Pool {
	result := &sync.Pool{}
	result.New = func() interface{} {
		return NewBuffer(result)
	}
	return result
}

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
