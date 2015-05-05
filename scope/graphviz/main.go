package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/scope/xfer"
)

func main() {
	var (
		defaultProbes = fmt.Sprintf("localhost:%d", xfer.ProbePort)
		probes        = flag.String("probes", defaultProbes, "list of probe endpoints, comma separated")
		batch         = flag.Duration("batch", 1*time.Second, "batch interval")
		window        = flag.Duration("window", 15*time.Second, "window")
		listen        = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
	)
	flag.Parse()

	xfer.MaxBackoff = 10 * time.Second
	c := xfer.NewCollector(strings.Split(*probes, ","), *batch)
	defer c.Stop()
	lifo := NewReportLIFO(c, *window)
	defer lifo.Stop()

	http.Handle("/svg", handleSVG(lifo))
	http.Handle("/txt", handleTXT(lifo))
	http.Handle("/", http.HandlerFunc(handleHTML))

	irq := interrupt()
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, nil))
		irq <- syscall.SIGINT
	}()
	<-irq
	log.Printf("shutting down")
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
