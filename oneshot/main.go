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

	"github.com/alicebob/cello/report"
	"github.com/alicebob/cello/xfer"
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
	c := xfer.NewCollector(strings.Split(*probes, ","), 1*time.Second)
	defer c.Stop()

	report := report.NewReport()
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
