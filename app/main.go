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
	"strings"
	"syscall"
	"time"

	"github.com/weaveworks/scope/xfer"
)

var (
	// Set at buildtime.
	version = "dev"

	// Set at runtime.
	uniqueID = "0"
)

func main() {
	var (
		window       = flag.Duration("window", 15*time.Second, "window")
		listen       = flag.String("http.address", ":"+strconv.Itoa(xfer.AppPort), "webserver listen address")
		logPrefix    = flag.String("log.prefix", "<app>", "prefix for each log line")
		printVersion = flag.Bool("version", false, "print version number and exit")
	)
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if !strings.HasSuffix(*logPrefix, " ") {
		*logPrefix += " "
	}
	log.SetPrefix(*logPrefix)

	defer log.Print("app exiting")

	rand.Seed(time.Now().UnixNano())
	uniqueID = strconv.FormatInt(rand.Int63(), 16)
	log.Printf("app starting, version %s, ID %s", version, uniqueID)

	c := xfer.NewCollector(*window)
	http.Handle("/", Router(c))
	go func() {
		log.Printf("listening on %s", *listen)
		log.Print(http.ListenAndServe(*listen, nil))
	}()
	log.Printf("%s", <-interrupt())
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
