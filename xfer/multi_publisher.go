package xfer

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

// MultiPublisher implements publisher over a collection of heterogeneous
// targets. See documentation of each method to understand the semantics.
type MultiPublisher struct {
	mtx     sync.Mutex
	factory func(endpoint string) (string, Publisher, error)
	sema    semaphore
	list    []tuple
}

// NewMultiPublisher returns a new MultiPublisher ready for use.
func NewMultiPublisher(factory func(endpoint string) (string, Publisher, error)) *MultiPublisher {
	return &MultiPublisher{
		factory: factory,
		sema:    newSemaphore(maxConcurrentGET),
	}
}

type tuple struct {
	publisher Publisher
	target    string // DNS name
	endpoint  string // IP addr
	id        string // unique ID from app
	err       error  // if factory failed
}

const maxConcurrentGET = 10

// Set declares that the target (DNS name) resolves to the provided endpoints
// (IPs), and that we want to publish to each of those endpoints. Set replaces
// any existing publishers to the given target. Set invokes the factory method
// to convert each endpoint to a publisher, and to get the remote receiver's
// unique ID.
func (p *MultiPublisher) Set(target string, endpoints []string) {
	// Convert endpoints to publishers.
	c := make(chan tuple, len(endpoints))
	for _, endpoint := range endpoints {
		go func(endpoint string) {
			p.sema.p()
			defer p.sema.v()
			id, publisher, err := p.factory(endpoint)
			c <- tuple{publisher, target, endpoint, id, err}
		}(endpoint)
	}
	list := make([]tuple, 0, len(p.list)+len(endpoints))
	for i := 0; i < cap(c); i++ {
		t := <-c
		if t.err != nil {
			log.Printf("multi-publisher set: %s (%s): %v", t.target, t.endpoint, t.err)
			continue
		}
		list = append(list, t)
	}

	// Copy all other tuples over to the new list.
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.list = p.appendFilter(list, func(t tuple) bool { return t.target != target })
}

// Delete removes all endpoints that match the given target.
func (p *MultiPublisher) Delete(target string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.list = p.appendFilter([]tuple{}, func(t tuple) bool { return t.target != target })
}

// Publish implements Publisher by publishing the reader to all of the
// underlying publishers sequentially. To do that, it needs to drain the
// reader, and recreate new readers for each publisher. Note that it will
// publish to one endpoint for each unique ID. Failed publishes don't count.
func (p *MultiPublisher) Publish(r io.Reader) error {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	var (
		ids  = map[string]struct{}{}
		errs = []string{}
	)

	p.mtx.Lock()
	defer p.mtx.Unlock()

	for _, t := range p.list {
		if _, ok := ids[t.id]; ok {
			continue
		}
		if err := t.publisher.Publish(bytes.NewReader(buf)); err != nil {
			errs = append(errs, err.Error())
			continue
		}
		ids[t.id] = struct{}{} // sent already
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Stop invokes stop on all underlying publishers and removes them.
func (p *MultiPublisher) Stop() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	for _, t := range p.list {
		t.publisher.Stop()
	}
	p.list = []tuple{}
}

func (p *MultiPublisher) appendFilter(list []tuple, f func(tuple) bool) []tuple {
	for _, t := range p.list {
		if !f(t) {
			t.publisher.Stop()
			continue
		}
		list = append(list, t)
	}
	return list
}

type semaphore chan struct{}

func newSemaphore(n int) semaphore {
	c := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		c <- struct{}{}
	}
	return semaphore(c)
}
func (s semaphore) p() { <-s }
func (s semaphore) v() { s <- struct{}{} }
