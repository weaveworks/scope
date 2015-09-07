package xfer

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/weaveworks/scope/report"
)

// Publisher is something which can send a report to a remote collector.
type Publisher interface {
	Publish(report.Report) error
	Stop()
}

// HTTPPublisher publishes reports by POST to a fixed endpoint.
type HTTPPublisher struct {
	url   string
	token string
	id    string
}

// ScopeProbeIDHeader is the header we use to carry the probe's unique ID. The
// ID is currently set to the probe's hostname. It's designed to deduplicate
// reports from the same probe to the same receiver, in case the probe is
// configured to publish to multiple receivers that resolve to the same app.
const ScopeProbeIDHeader = "X-Scope-Probe-ID"

// NewHTTPPublisher returns an HTTPPublisher ready for use.
func NewHTTPPublisher(target, token, id string) (*HTTPPublisher, error) {
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if u.Path == "" {
		u.Path = "/api/report"
	}
	return &HTTPPublisher{
		url:   u.String(),
		token: token,
		id:    id,
	}, nil
}

// Publish publishes the report to the URL.
func (p HTTPPublisher) Publish(rpt report.Report) error {
	gzbuf := bytes.Buffer{}
	gzwriter := gzip.NewWriter(&gzbuf)

	if err := gob.NewEncoder(gzwriter).Encode(rpt); err != nil {
		return err
	}
	gzwriter.Close() // otherwise the content won't get flushed to the output stream

	req, err := http.NewRequest("POST", p.url, &gzbuf)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", AuthorizationHeader(p.token))
	req.Header.Set(ScopeProbeIDHeader, p.id)
	req.Header.Set("Content-Encoding", "gzip")
	// req.Header.Set("Content-Type", "application/binary") // TODO: we should use http.DetectContentType(..) on the gob'ed

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

// Stop implements Publisher
func (p HTTPPublisher) Stop() {}

// AuthorizationHeader returns a value suitable for an HTTP Authorization
// header, based on the passed token string.
func AuthorizationHeader(token string) string {
	return fmt.Sprintf("Scope-Probe token=%s", token)
}

// BackgroundPublisher is a publisher which does the publish asynchronously.
// It will only do one publish at once; if there is an ongoing publish,
// concurrent publishes are dropped.
type BackgroundPublisher struct {
	publisher Publisher
	reports   chan report.Report
}

// NewBackgroundPublisher creates a new BackgroundPublisher with the given publisher
func NewBackgroundPublisher(p Publisher) *BackgroundPublisher {
	result := &BackgroundPublisher{
		publisher: p,
		reports:   make(chan report.Report),
	}
	go result.loop()
	return result
}

func (b *BackgroundPublisher) loop() {
	for r := range b.reports {
		if err := b.publisher.Publish(r); err != nil {
			log.Printf("Error publishing: %v", err)
		}
	}
}

// Publish implements Publisher
func (b *BackgroundPublisher) Publish(r report.Report) error {
	select {
	case b.reports <- r:
		return nil
	default:
		return fmt.Errorf("Dropping report - can't push fast enough")
	}
}

// Stop implements Publisher
func (b *BackgroundPublisher) Stop() {
	close(b.reports)
	b.publisher.Stop()
}

// MultiPublisher implements Publisher over a set of publishers.
type MultiPublisher struct {
	mtx     sync.RWMutex
	factory func(string) (Publisher, error)
	m       map[string]Publisher
}

// NewMultiPublisher returns a new MultiPublisher ready for use. The factory
// should be e.g. NewHTTPPublisher, except you need to curry it over the
// probe token.
func NewMultiPublisher(factory func(string) (Publisher, error)) *MultiPublisher {
	return &MultiPublisher{
		factory: factory,
		m:       map[string]Publisher{},
	}
}

// Add allows additional targets to be added dynamically. It will dedupe
// identical targets. TODO we have no good mechanism to remove.
func (p *MultiPublisher) Add(target string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if _, ok := p.m[target]; ok {
		return
	}

	publisher, err := p.factory(target)
	if err != nil {
		log.Printf("multi-publisher: %v", err)
		return
	}

	p.m[target] = publisher
}

// Publish implements Publisher by emitting the report to all publishers.
func (p *MultiPublisher) Publish(rpt report.Report) error {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	var errs []string
	for _, publisher := range p.m {
		if err := publisher.Publish(rpt); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

// Stop implements Publisher
func (p *MultiPublisher) Stop() {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	for _, publisher := range p.m {
		publisher.Stop()
	}
}
