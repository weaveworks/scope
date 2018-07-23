package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	metric, err := speed.NewPCPCounter(
		0,
		"counter",
		"A Simple Metric",
	)
	if err != nil {
		log.Fatal("Could not create counter, error: ", err)
	}

	client, err := speed.NewPCPClient("singletoncounter")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	err = client.Register(metric)
	if err != nil {
		log.Fatal("Could not register metric, error: ", err)
	}

	client.MustStart()
	defer client.MustStop()

	fmt.Println("The metric should be visible as mmv.singletoncounter.counter")
	for i := 0; i < *timelimit; i++ {
		metric.Up()
		time.Sleep(time.Second)
	}
}
