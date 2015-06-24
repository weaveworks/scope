package xfer

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestCollector(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Swap out ticker
	publish := make(chan time.Time)
	oldTick := tick
	tick = func(time.Duration) <-chan time.Time { return publish }
	defer func() { tick = oldTick }()

	// Build a collector
	collector := NewCollector(time.Second, "id")
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
	r := report.Report{
		Address: report.Topology{
			NodeMetadatas: report.NodeMetadatas{
				report.MakeAddressNodeID("a", "b"): report.NodeMetadata{},
			},
		},
	}

	reports <- r
	poll(t, 100*time.Millisecond, func() bool {
		return len(concreteCollector.peek().Address.NodeMetadatas) == 1
	}, "missed the report")

	go func() { publish <- time.Now() }()
	collected := <-collector.Reports()
	if reflect.DeepEqual(r, collected) {
		t.Errorf(test.Diff(r, collected))
	}

	collector.Remove(addr)
	collector.Remove(addr) // test duplicate case
}

func TestCollectorQuitWithActiveConnections(t *testing.T) {
	c := NewCollector(time.Second, "id")
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
