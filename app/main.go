package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/weaveworks/scope/xfer"
)

// Set during buildtime.
var version = "dev"

func main() {
	var (
		defaultProbes = []string{fmt.Sprintf("localhost:%d", xfer.ProbePort), fmt.Sprintf("scope.weave.local:%d", xfer.ProbePort)}
		batch         = flag.Duration("batch", 1*time.Second, "batch interval")
		window        = flag.Duration("window", 15*time.Second, "window")
		listen        = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		printVersion  = flag.Bool("version", false, "print version number and exit")
	)
	flag.Parse()
	probes := append(defaultProbes, flag.Args()...)

	if *printVersion {
		fmt.Println(version)
		return
	}

	rand.Seed(time.Now().UnixNano())
	id := strconv.FormatInt(rand.Int63(), 16)
	log.Printf("app starting, version %s, id %s", version, id)

	// Collector deals with the probes, and generates merged reports.
	c := xfer.NewCollector(*batch, id)
	defer c.Stop()

	r := newStaticResolver(probes, c.Add)
	defer r.Stop()

	lifo := NewReportLIFO(c, *window)
	defer lifo.Stop()

	http.Handle("/", Router(lifo))
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
