package xfer

import (
	"encoding/gob"
	"log"
	"net"
	"sync"
	"time"

	"github.com/weaveworks/scope/scope/report"
)

const (
	connectTimeout = 10 * time.Second
	initialBackoff = 2 * time.Second
)

var (
	// MaxBackoff is the maximum time between connect retries.
	MaxBackoff = 2 * time.Minute // externally configurable.
)

// Collector connects to probes over TCP and merges reports published by those
// probes into a single one.
type Collector struct {
	in     chan report.Report
	out    chan report.Report
	add    chan string
	remove chan string
	quit   chan chan struct{}
}

// NewCollector starts the report collector.
func NewCollector(ips []string, batchTime time.Duration) *Collector {
	c := &Collector{
		in:     make(chan report.Report),
		out:    make(chan report.Report),
		add:    make(chan string),
		remove: make(chan string),
		quit:   make(chan chan struct{}),
	}

	go c.loop(ips, batchTime)

	return c
}

func (c *Collector) loop(ips []string, batchTime time.Duration) {
	var (
		tick    = time.Tick(batchTime)
		current = report.NewReport()
		addrs   = map[string]chan struct{}{}
		wg      = &sync.WaitGroup{} // individual collector goroutines
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

	for _, ip := range ips {
		add(ip)
	}

	for {
		select {
		case <-tick:
			c.out <- current
			current = report.NewReport()

		case r := <-c.in:
			current.Merge(r)

		case ip := <-c.add:
			add(ip)

		case ip := <-c.remove:
			remove(ip)

		case q := <-c.quit:
			for _, q := range addrs {
				close(q)
			}
			wg.Wait()
			close(q)
			return
		}
	}
}

// Stop shuts down a collector and all connections to probes.
func (c *Collector) Stop() {
	q := make(chan struct{})
	c.quit <- q
	<-q
}

// AddAddress adds the passed IP to the collector, and starts (trying to)
// collect reports from the remote Publisher.
func (c *Collector) AddAddress(ip string) {
	c.add <- ip
}

// RemoveAddress removes the passed IP from the collector, and stops
// collecting reports from the remote Publisher.
func (c *Collector) RemoveAddress(ip string) {
	c.remove <- ip
}

// Reports returns the channel where aggregate reports are sent.
func (c *Collector) Reports() <-chan report.Report {
	return c.out
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
			//log.Printf("collector: got a report from %v", ip)

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
