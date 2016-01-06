// Publish a fixed report.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
	"github.com/weaveworks/scope/report"
)

func main() {
	var (
		publish         = flag.String("publish", fmt.Sprintf("localhost:%d", xfer.AppPort), "publish target")
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
	)
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("usage: fixprobe [--args] report.json")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	var fixedReport report.Report
	if err := json.NewDecoder(f).Decode(&fixedReport); err != nil {
		log.Fatal(err)
	}
	f.Close()

	client, err := appclient.NewAppClient(appclient.ProbeConfig{
		Token:    "fixprobe",
		ProbeID:  "fixprobe",
		Insecure: false,
	}, *publish, *publish, nil)
	if err != nil {
		log.Fatal(err)
	}

	rp := appclient.NewReportPublisher(client)
	for range time.Tick(*publishInterval) {
		rp.Publish(fixedReport)
	}
}
