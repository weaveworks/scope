package main

import (
	"flag"
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
		window = flag.Duration("window", 15*time.Second, "window")
		listen = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
	)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	id := strconv.FormatInt(rand.Int63(), 16)
	log.Printf("app starting, version %s, ID %s", version, id)

	c := xfer.NewCollector(*window)
	http.Handle("/", Router(c))
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
