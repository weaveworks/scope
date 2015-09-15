package xfer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// Sender takes ownership of ReadClosers and tries to send them to a remote
// receiver. Senders guarantee that the ReadCloser will be closed, even if it
// isn't successfully sent.
type Sender interface {
	Send(io.ReadCloser) // must never block or fail
	Stop()
}

// HTTPSender implements Sender by sending the most recent ReadCloser to a
// remote target in the body of an HTTP POST. If the remote target is broken
// or slow, HTTPSender will drop (and close) older ReadClosers in favor of
// newer ones.
type HTTPSender struct {
	target    string
	authToken string
	probeID   string
	recv      chan io.ReadCloser // from clients
	send      chan io.ReadCloser // to remote
	quit      chan struct{}
}

// NewHTTPSender returns a usable HTTP-based Sender.
func NewHTTPSender(target, authToken, probeID string) Sender {
	s := &HTTPSender{
		target:    target, // TODO(pb): sanitize.URL("http://", AppPort, "/api/report")(target),
		authToken: authToken,
		probeID:   probeID,
		recv:      make(chan io.ReadCloser),
		send:      make(chan io.ReadCloser),
		quit:      make(chan struct{}),
	}
	go s.recvLoop()
	go s.sendLoop()
	return s
}

// Send implements Sender. It is nonblocking.
func (s *HTTPSender) Send(rc io.ReadCloser) {
	s.recv <- rc // will (must) never block
}

// Stop terminates the sender.
func (s *HTTPSender) Stop() {
	close(s.quit)
}

func (s *HTTPSender) recvLoop() {
	var (
		rc   io.ReadCloser      // the most recent ReadCloser we got
		send chan io.ReadCloser // nil until we get the first ReadCloser
	)
	for {
		select {
		case rc0 := <-s.recv:
			if rc != nil {
				rc.Close()
				rc = nil
				log.Printf("dropped a report to %s, slow app?", s.target)
			}
			rc = rc0
			send = s.send // we should send it on

		case send <- rc:
			send = nil // send OK, wait for the next rc

		case <-s.quit:
			if rc != nil {
				rc.Close()
				rc = nil
			}
			return
		}
	}
}

func (s *HTTPSender) sendLoop() {
	for {
		select {
		case rc := <-s.send:
			// If this blocks, it's OK.
			if err := s.post(rc); err != nil {
				log.Printf("%s: %v", s.target, err)
				// TODO(pb): do we still need backoff?
			}

		case <-s.quit:
			return
		}
	}
}

func (s *HTTPSender) post(rc io.ReadCloser) error {
	defer rc.Close()

	req, err := http.NewRequest("POST", s.target, rc)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", AuthorizationHeader(s.authToken))
	req.Header.Set(ScopeProbeIDHeader, s.probeID)
	req.Header.Set("Content-Encoding", "gzip")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
// ID is currently set to the probe's hostname. It's designed to deduplicate
// reports from the same probe to the same receiver, in case the probe is
// configured to publish to multiple receivers that resolve to the same app.
const ScopeProbeIDHeader = "X-Scope-Probe-ID"

// AuthorizationHeader returns a value suitable for an HTTP Authorization
// header, based on the passed token string.
func AuthorizationHeader(token string) string {
	return fmt.Sprintf("Scope-Probe token=%s", token)
}

// MultiSender implements Sender by demuxing ReadClosers to multiple senders.
// New senders are constructed via the factory function, and added via the Add
// method.
type MultiSender struct {
	mtx     sync.RWMutex
	factory func(string) (Sender, error)
	senders map[string]Sender
	pool    *sync.Pool
}

// NewMultiSender returns a new MultiSender, ready for use.
func NewMultiSender(factory func(string) (Sender, error)) *MultiSender {
	return &MultiSender{
		factory: factory,
		senders: map[string]Sender{},
		pool:    &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }},
	}
}

// Add adds a new sender to the multi-sender, if it doesn't already exist.
func (ms *MultiSender) Add(target string) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	if _, ok := ms.senders[target]; ok {
		return
	}

	sender, err := ms.factory(target)
	if err != nil {
		log.Printf("multi-sender: %s: %v", target, err)
		return
	}

	ms.senders[target] = sender
}

// Send implements Sender by copying the contents of the rc to each sender.
// Once the contents of the passed ReadCloser is copied N times, it's closed.
//
// Send uses a buffer pool to reduce GC pressure, but does burn CPU to copy.
// This could be optimized with e.g. reference counting.
func (ms *MultiSender) Send(rc io.ReadCloser) {
	defer rc.Close() // end-of-line for this one

	ms.mtx.RLock()
	defer ms.mtx.RUnlock()

	// Get a buffer for each sender.
	rwcs := make([]io.ReadWriteCloser, len(ms.senders))
	for i := range rwcs {
		buf := ms.pool.Get().(*bytes.Buffer)
		buf.Reset()
		rwcs[i] = &pooledBuffer{buf, ms.pool}
	}

	// Copy the rc into the buffers.
	writers := make([]io.Writer, len(rwcs))
	for i := range writers {
		writers[i] = rwcs[i]
	}
	if _, err := io.Copy(io.MultiWriter(writers...), rc); err != nil {
		log.Printf("multi-sender: during demux: %v", err)
		return
	}

	// Send each buffer. These can't (won't) block.
	i := 0
	for _, sender := range ms.senders {
		sender.Send(rwcs[i])
		i++
	}
}

// Stop implements Sender by stopping all of the underlying senders.
func (ms *MultiSender) Stop() {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	for _, sender := range ms.senders {
		sender.Stop()
	}
}

// pooledBuffer is a bytes.Buffer that came out of a sync.Pool. The Close
// method returns the buffer to the pool.
//
// Only use pooledBuffer when you need to pass ownership of a buffer away from
// its creation context. For example, the multi-sender hands a buffer to the
// HTTP sender, and allows the HTTP sender to potentially drop the buffer.
// Otherwise, use a plain sync.Pool.
type pooledBuffer struct {
	*bytes.Buffer
	parent *sync.Pool
}

func (pb *pooledBuffer) Close() error {
	if pb.Buffer == nil {
		return errDoubleClose
	}
	pb.parent.Put(pb.Buffer)
	pb.Buffer = nil // gone
	return nil
}

var errDoubleClose = errors.New("double-close on a pooled buffer")
