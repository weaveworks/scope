package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func main() {
	var (
		defaultProbes = fmt.Sprintf("localhost:%d", xfer.ProbePort)
		probes        = flag.String("probes", defaultProbes, "list of probe endpoints, comma separated")
	)
	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Collector deals with the probes, and generates merged reports.
	xfer.MaxBackoff = 1 * time.Second
	c := xfer.NewCollector(1 * time.Second)
	for _, addr := range strings.Split(*probes, ",") {
		c.Add(addr)
	}
	defer c.Stop()

	report := report.MakeReport()
	irq := interrupt()
OUTER:
	for {
		select {
		case r := <-c.Reports():
			report.Merge(r)
		case <-irq:
			break OUTER
		}
	}

	b, err := json.Marshal(report)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
