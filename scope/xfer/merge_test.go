package xfer_test

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/weaveworks/scope/scope/report"
	"github.com/weaveworks/scope/scope/xfer"
)

func TestMerge(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	var (
		p1Addr = "localhost:7888"
		p2Addr = "localhost:7889"
	)

	p1, err := xfer.NewTCPPublisher(p1Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer p1.Close()

	p2, err := xfer.NewTCPPublisher(p2Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()

	batchTime := 100 * time.Millisecond
	c := xfer.NewCollector([]string{p1Addr, p2Addr}, batchTime)
	defer c.Stop()
	time.Sleep(batchTime / 10) // connect

	{
		r := report.NewReport()
		r.HostMetadatas["p1"] = report.HostMetadata{Hostname: "test1"}
		p1.Publish(r)
	}
	{
		r := report.NewReport()
		r.HostMetadatas["p2"] = report.HostMetadata{Hostname: "test2"}
		p2.Publish(r)
	}

	success := make(chan struct{})
	go func() {
		defer close(success)
		for r := range c.Reports() {
			if r.HostMetadatas["p1"].Hostname != "test1" {
				continue
			}
			if r.HostMetadatas["p2"].Hostname != "test2" {
				continue
			}
			return
		}
	}()

	select {
	case <-success:
	case <-time.After(batchTime):
		t.Errorf("collector didn't capture both reports")
	}
}
