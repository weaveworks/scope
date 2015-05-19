package xfer_test

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func TestTCPPublisher(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Choose a port
	port, err := getFreePort()
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
	conn, err := net.Dial("tcp4", "127.0.0.1"+port)
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

func TestPublisherClosesDuplicateConnections(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	// Choose a port
	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	// Start a publisher
	p, err := xfer.NewTCPPublisher(port)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// Connect a listener
	conn, err := net.Dial("tcp4", "127.0.0.1"+port)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(time.Millisecond)

	// Try to connect the same listener
	dupconn, err := net.Dial("tcp4", "127.0.0.1"+port)
	if err != nil {
		t.Fatal(err)
	}
	defer dupconn.Close()

	// Publish a message
	p.Publish(report.Report{})

	// The first listener should receive it
	var r report.Report
	if err := gob.NewDecoder(conn).Decode(&r); err != nil {
		t.Fatal(err)
	}

	// The duplicate listener should have an error
	if err := gob.NewDecoder(dupconn).Decode(&r); err == nil {
		t.Errorf("expected error, got none")
	} else {
		t.Logf("dupconn got expected error: %v", err)
	}
}

func getFreePort() (string, error) {
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		return "", fmt.Errorf("Listen: %v", err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return "", fmt.Errorf("SplitHostPort(%s): %v", ln.Addr().String(), err)
	}
	return ":" + port, nil
}
