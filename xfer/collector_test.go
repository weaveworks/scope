package xfer

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
)

func TestCollector(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Swap out ticker
	publish := make(chan time.Time)
	oldTick := tick
	tick = func(time.Duration) <-chan time.Time { return publish }
	defer func() { tick = oldTick }()

	// Build a collector
	collector := NewCollector(time.Second)
	defer collector.Stop()

	concreteCollector, ok := collector.(*realCollector)
	if !ok {
		t.Fatal("type assertion failure")
	}

	// Build a test publisher
	reports := make(chan interface{})
	ln := testPublisher(t, reports)
	defer ln.Close()

	// Connect the collector to the test publisher
	addr := ln.Addr().String()
	collector.Add(addr)
	collector.Add(addr) // test duplicate case
	runtime.Gosched()   // make sure it connects

	// Push a report through everything
	reports <- report.Report{Host: report.Topology{NodeMetadatas: report.NodeMetadatas{"a": report.NodeMetadata{}}}}
	poll(t, time.Millisecond, func() bool { return len(concreteCollector.peek().Host.NodeMetadatas) == 1 }, "missed the report")
	go func() { publish <- time.Now() }()
	if want, have := 1, len((<-collector.Reports()).Host.NodeMetadatas); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	collector.Remove(addr)
	collector.Remove(addr) // test duplicate case
}

func TestCollectorQuitWithActiveConnections(t *testing.T) {
	c := NewCollector(time.Second)
	c.Add("1.2.3.4:56789")
	c.Stop()
}

func testPublisher(t *testing.T, input <-chan interface{}) net.Listener {
	addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Log(err)
			return
		}
		defer conn.Close()
		for {
			enc := gob.NewEncoder(conn)
			for v := range input {
				if err := enc.Encode(v); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()
	return ln
}

func poll(t *testing.T, d time.Duration, condition func() bool, msg string) {
	deadline := time.Now().Add(d)
	for {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		if condition() {
			return
		}
		time.Sleep(d / 10)
	}
}
