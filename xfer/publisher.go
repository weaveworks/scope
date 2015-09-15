package xfer

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"io"
	"io/ioutil"
	"sync"

	"github.com/weaveworks/scope/report"
)

// Publisher publishes a report.Report to a remote collector.
type Publisher interface {
	Publish(report.Report) error
}

// ReportEncoder is used to serialize reports.
type ReportEncoder func(dst io.Writer, src report.Report) error

// GzipGobEncoder is the default report encoder.
func GzipGobEncoder(dst io.Writer, src report.Report) error {
	gzw := gzip.NewWriter(dst)
	if err := gob.NewEncoder(gzw).Encode(src); err != nil {
		return err
	}
	return gzw.Close() // required to flush
}

// JSONEncoder is a debug report encoder.
func JSONEncoder(dst io.Writer, src report.Report) error {
	return json.NewEncoder(dst).Encode(src)
}

// SendingPublisher publishes reports by serializing them to a Sender.
// It serializes into buffers managed by a sync.Pool to reduce GC pressure.
type SendingPublisher struct {
	enc    ReportEncoder
	sender Sender
	pool   *sync.Pool
}

// NewSendingPublisher returns a new SendingPublisher ready to use.
func NewSendingPublisher(enc ReportEncoder, s Sender) Publisher {
	return SendingPublisher{
		enc:    enc,
		sender: s,
		pool:   &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
	}
}

// Publish implements Publisher by serializing the report and forwarding it
// to the sender.
func (sp SendingPublisher) Publish(rpt report.Report) error {
	buf := sp.pool.Get().(*bytes.Buffer)
	defer sp.pool.Put(buf)

	buf.Reset()
	if err := sp.enc(buf, rpt); err != nil {
		return err
	}

	// When send returns, the sender is guaranteed to have called Close and be
	// done with the buf. But we don't leverage that fact (via pooledBuffer)
	// because we don't need to: we're able to return the buf to the pool in
	// our context (via the defer) so that's what we do.
	sp.sender.Send(ioutil.NopCloser(buf))
	return nil
}
