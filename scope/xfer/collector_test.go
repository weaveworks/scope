package xfer_test

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"

	"github.com/weaveworks/scope/scope/report"
	"github.com/weaveworks/scope/scope/xfer"
)

func TestCollector(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Build the address
	port := ":12345"
	addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1"+port)
	if err != nil {
		t.Fatal(err)
	}

	// Start a raw publisher
	ln, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Accept one connection, write one report
	data := make(chan []byte)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		if _, err := conn.Write(<-data); err != nil {
			t.Error(err)
			return
		}
	}()

	// Start a collector
	batchTime := 10 * time.Millisecond
	c := xfer.NewCollector([]string{"127.0.0.1" + port}, batchTime)
	gate := make(chan struct{})
	go func() { <-c.Reports(); c.Stop(); close(gate) }()

	// Publish a message
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(report.Report{}); err != nil {
		t.Fatal(err)
	}
	data <- buf.Bytes()

	// Check it was collected and forwarded
	select {
	case <-gate:
	case <-time.After(2 * batchTime):
		t.Errorf("timeout waiting for report")
	}
}
