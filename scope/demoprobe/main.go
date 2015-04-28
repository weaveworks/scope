package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/alicebob/cello/xfer"
)

func main() {
	var (
		version         = flag.Bool("version", false, "print version number and exit")
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
		listen          = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
		hostCount       = flag.Int("hostcount", 10, "Number of demo hosts to generate")
	)
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	// -version flag:
	if *version {
		fmt.Printf("unstable\n")
		return
	}

	publisher, err := xfer.NewTCPPublisher(*listen)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()
	go func() {
		for {
			publisher.Publish(DemoReport(*hostCount))
			time.Sleep(*publishInterval)
		}
	}()

	log.Printf("%s", <-interrupt())
	log.Printf("Shutting down...")
}

func interrupt() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}
