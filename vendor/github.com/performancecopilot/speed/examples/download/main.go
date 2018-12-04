// This is an example of using composite units
// It makes successive downloads of the linux source code
// and reports the download speed as well as the download time
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/performancecopilot/speed"
)

type sink int

func (s sink) Write(data []byte) (int, error) {
	return len(data), nil
}

const url = "https://codeload.github.com/suyash/ulid/zip/master"

func main() {
	client, err := speed.NewPCPClient("download")
	if err != nil {
		log.Fatal(err)
	}

	downloadSpeed, err := speed.NewPCPSingletonMetric(
		float64(0),
		"download_speed",
		speed.DoubleType,
		speed.InstantSemantics,
		speed.MegabyteUnit.Time(speed.SecondUnit, -1),
	)

	if err != nil {
		log.Fatal(err)
	}

	timer, err := speed.NewPCPTimer("download_time", speed.SecondUnit)
	if err != nil {
		log.Fatal(err)
	}

	client.MustRegister(timer)
	client.MustRegister(downloadSpeed)

	client.MustStart()
	defer client.MustStop()

	for {
		timer.Reset()
		run(timer, downloadSpeed)
	}
}

func run(timer *speed.PCPTimer, downloadSpeed *speed.PCPSingletonMetric) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	var s sink
	err = timer.Start()
	if err != nil {
		log.Fatal(err)
	}

	n, err := io.Copy(s, res.Body)
	if err != nil {
		log.Fatal(err)
	}

	elapsed, err := timer.Stop()
	if err != nil {
		log.Fatal(err)
	}

	downloadSpeed.Set(float64(n) / (1024 * 1024 * float64(elapsed)))
	log.Println("downloaded", n, "bytes in", elapsed, "seconds")
}
