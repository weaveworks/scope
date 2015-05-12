package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/xfer"
)

func main() {
	var (
		defaultProbes = []string{fmt.Sprintf("localhost:%d", xfer.ProbePort), fmt.Sprintf("scope.weave.local:%d", xfer.ProbePort)}
		logfile       = flag.String("log", "stderr", "stderr, syslog, or filename")
		probes        = flag.String("probes", strings.Join(defaultProbes, ","), "list of probe endpoints, comma separated")
		batch         = flag.Duration("batch", 1*time.Second, "batch interval")
		window        = flag.Duration("window", 15*time.Second, "window")
		pidfile       = flag.String("pidfile", "", "write PID file")
		listen        = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
	)
	flag.Parse()

	switch *logfile {
	case "stderr":
		break // by default

	case "syslog":
		w, err := syslog.New(syslog.LOG_INFO, "cello-app")
		if err != nil {
			log.Print(err)
			return
		}
		defer w.Close()
		log.SetFlags(0)
		log.SetOutput(w)

	default: // file
		f, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if *pidfile != "" {
		err := ioutil.WriteFile(*pidfile, []byte(fmt.Sprint(os.Getpid())), 0644)
		if err != nil {
			log.Print(err)
			return
		}
		defer os.Remove(*pidfile)
	}

	log.Printf("starting")

	// Collector deals with the probes, and generates merged reports.
	xfer.MaxBackoff = 10 * time.Second
	c := xfer.NewCollector(*batch)
	defer c.Stop()

	r := NewResolver(strings.Split(*probes, ","), c.AddAddress)
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
