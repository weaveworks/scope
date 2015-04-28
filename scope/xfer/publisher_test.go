package xfer_test

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"

	"github.com/alicebob/cello/report"
	"github.com/alicebob/cello/xfer"
)

func TestTCPPublisher(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Build the address
	port := ":12345"
	addr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1"+port)
	if err != nil {
		t.Fatal(err)
	}

	// Start a publisher
	p, err := xfer.NewTCPPublisher(port)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// Start a raw listener
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond)

	// Publish a message
	p.Publish(report.Report{})

	// Check it was received
	var r report.Report
	if err := gob.NewDecoder(conn).Decode(&r); err != nil {
		t.Fatal(err)
	}
}
