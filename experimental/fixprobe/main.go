// Publish a fixed report.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/xfer"
)

func main() {
	var (
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
		listenAddress   = flag.String("listen", ":"+strconv.Itoa(xfer.ProbePort), "listen address")
	)
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("usage: fixprobe [--args] report.json\n")
		return
	}
	fixture := flag.Arg(0)

	f, err := os.Open(fixture)
	if err != nil {
		fmt.Printf("json error: %v\n", err)
		return
	}
	var fixedReport report.Report
	if err := json.NewDecoder(f).Decode(&fixedReport); err != nil {
		fmt.Printf("json error: %v\n", err)
		return
	}

	publisher, err := xfer.NewTCPPublisher(*listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	log.Printf("listening on %s", *listenAddress)

	for range time.Tick(*publishInterval) {
		publisher.Publish(fixedReport)
	}
}
