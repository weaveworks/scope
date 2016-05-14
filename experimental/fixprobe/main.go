// Publish a fixed report.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/ugorji/go/codec"

	"$GITHUB_URI/common/xfer"
	"$GITHUB_URI/probe/appclient"
	"$GITHUB_URI/report"
	"$GITHUB_URI/test/fixture"
)

func main() {
	var (
		publish         = flag.String("publish", fmt.Sprintf("localhost:%d", xfer.AppPort), "publish target")
		publishInterval = flag.Duration("publish.interval", 1*time.Second, "publish (output) interval")
		publishToken    = flag.String("publish.token", "fixprobe", "publish token, for if we are talking to the service")
		publishID       = flag.String("publish.id", "fixprobe", "publisher ID used to identify publishers")
		useFixture      = flag.Bool("fixture", false, "Use the embedded fixture report.")
	)
	flag.Parse()

	if len(flag.Args()) != 1 && !*useFixture {
		log.Fatal("usage: fixprobe [--args] report.json")
	}

	var fixedReport report.Report
	if *useFixture {
		fixedReport = fixture.Report
	} else {
		b, err := ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}

		decoder := codec.NewDecoderBytes(b, &codec.JsonHandle{})
		if err := decoder.Decode(&fixedReport); err != nil {
			log.Fatal(err)
		}
	}

	client, err := appclient.NewAppClient(appclient.ProbeConfig{
		Token:    *publishToken,
		ProbeID:  *publishID,
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
