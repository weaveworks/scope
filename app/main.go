package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/weaveworks/scope/xfer"
)

var version = "dev" // set at build time

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

	log.Printf("app version %s", version)

	xfer.MaxBackoff = 10 * time.Second
	c := xfer.NewCollector(*batch)
	defer c.Stop()

	r := newStaticResolver(probes, c.Add)
	defer r.Stop()

	reporter := newLIFOReporter(c.Reports(), *window)
	defer reporter.Stop()

	errc := make(chan error)
	go func() {
		http.Handle("/", newRouter(reporter))
		log.Printf("listening on %s", *listen)
		errc <- http.ListenAndServe(*listen, nil)
	}()
	go func() {
		errc <- interrupt()
	}()
	log.Print(<-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
