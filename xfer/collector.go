package xfer

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
	"time"

	"github.com/weaveworks/scope/report"
)

const (
	connectTimeout = 10 * time.Second
	initialBackoff = 2 * time.Second
)

var (
	// MaxBackoff is the maximum time between connect retries.
	// It's exported so it's externally configurable.
	MaxBackoff = 2 * time.Minute

	// This is extracted out for mocking.
	tick = time.Tick
)

// Collector describes anything that can have addresses added and removed, and
// which produces reports that represent aggregate reports from all collected
// addresses.
type Collector interface {
	Add(string)
	Remove(string)
	Reports() <-chan report.Report
	Stop()
}

// realCollector connects to probes over TCP and merges reports published by those
// probes into a single one.
type realCollector struct {
	in     chan report.Report
	out    chan report.Report
	peekc  chan chan report.Report
	add    chan string
	remove chan string
	quit   chan struct{}
}

// NewCollector produces and returns a report collector.
func NewCollector(batchTime time.Duration) Collector {
	c := &realCollector{
		in:     make(chan report.Report),
		out:    make(chan report.Report),
		peekc:  make(chan chan report.Report),
		add:    make(chan string),
		remove: make(chan string),
		quit:   make(chan struct{}),
	}
	go c.loop(batchTime)
	return c
}

func (c *realCollector) loop(batchTime time.Duration) {
	var (
		tick    = tick(batchTime)
		current = report.MakeReport()
		addrs   = map[string]chan struct{}{}
		wg      = &sync.WaitGroup{} // per-address goroutines
	)

	add := func(ip string) {
		if _, ok := addrs[ip]; ok {
			return
		}
		addrs[ip] = make(chan struct{})
		wg.Add(1)
		go func(quit chan struct{}) {
			defer wg.Done()
			reportCollector(ip, c.in, quit)
		}(addrs[ip])
	}

	remove := func(ip string) {
		q, ok := addrs[ip]
		if !ok {
			return // hmm
		}
		close(q)
		delete(addrs, ip)
	}

	for {
		select {
		case <-tick:
			c.out <- current
			current = report.MakeReport()

		case pc := <-c.peekc:
			copy := report.MakeReport()
			copy.Merge(current)
			pc <- copy

		case r := <-c.in:
			current.Merge(r)

		case ip := <-c.add:
			add(ip)

		case ip := <-c.remove:
			remove(ip)

		case <-c.quit:
			for _, q := range addrs {
				close(q)
			}
			wg.Wait()
			return
		}
	}
}

// Add adds an address to be collected from.
func (c *realCollector) Add(addr string) {
	c.add <- addr
}

// Remove removes a previously-added address.
func (c *realCollector) Remove(addr string) {
	c.remove <- addr
}

// Reports returns the report chan. It must be consumed by the client, or the
// collector will break.
func (c *realCollector) Reports() <-chan report.Report {
	return c.out
}

func (c *realCollector) peek() report.Report {
	pc := make(chan report.Report)
	c.peekc <- pc
	return <-pc
}

// Stop terminates the collector.
func (c *realCollector) Stop() {
	close(c.quit)
}

// reportCollector is the loop to connect to a single Probe. It'll keep
// running until the quit channel is closed.
func reportCollector(ip string, col chan<- report.Report, quit <-chan struct{}) {
	backoff := initialBackoff / 2
	for {
		backoff *= 2
		if backoff > MaxBackoff {
			backoff = MaxBackoff
		}

		select {
		default:
		case <-quit:
			return
		}

		log.Printf("dialing %v (backoff %v)", ip, backoff)

		conn, err := net.DialTimeout("tcp", ip, connectTimeout)
		if err != nil {
			log.Print(err)
			select {
			case <-time.After(backoff):
				continue
			case <-quit:
				return
			}
		}

		log.Printf("connected to %v", ip)

		go func() {
			<-quit
			log.Printf("closing %v collector", ip)
			conn.Close()
		}()

		// Connection accepted.
		dec := gob.NewDecoder(conn)
		for {
			var report report.Report
			err := dec.Decode(&report)
			// Don't complain of errors when shutting down.
			select {
			default:
			case <-quit:
				return
			}
			if err != nil {
				log.Printf("decode error: %v", err)
				break
			}

			select {
			case col <- report:
			case <-quit:
				return
			}

			// Reset the backoff iff we have a connection which works. This
			// prevents us from spamming probes with multiple addresses (since
			// the probe closes everything but a single connection).
			backoff = initialBackoff
		}

		// Prevent a 100% CPU loop when the probe is closing the
		// connection right away (which happens on a probe which already
		// has a client)
		select {
		case <-time.After(backoff):
		case <-quit:
			return
		}
	}
}
