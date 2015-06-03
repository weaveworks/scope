package xfer_test

import (
	"encoding/gob"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func TestTCPPublisher(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	addr := getFreeAddr(t)
	p, err := xfer.NewTCPPublisher(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// Connect a listener
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	runtime.Gosched()

	// Connect a duplicate listener
	dupconn, err := net.Dial("tcp4", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer dupconn.Close()

	// Publish a message
	p.Publish(report.Report{})

	// The first listener should receive it
	var r report.Report
	if err := gob.NewDecoder(conn).Decode(&r); err != nil {
		t.Error(err)
	}

	// The duplicate listener should have an error
	if err := gob.NewDecoder(dupconn).Decode(&r); err == nil {
		t.Errorf("expected error, got none")
	} else {
		t.Logf("dupconn got expected error: %v", err)
	}
}

func getFreeAddr(t *testing.T) string {
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort(%s): %v", ln.Addr().String(), err)
	}
	return "127.0.0.1:" + port
}
