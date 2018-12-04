package main

import (
	"fmt"
	"log"
	"os"

	"github.com/performancecopilot/speed"
)

func main() {
	metric, err := speed.NewPCPSingletonMetric(
		42,
		"simple.counter",
		speed.Int32Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"A Simple Metric",
		"This is a simple counter metric to demonstrate the speed API",
	)
	if err != nil {
		log.Fatal("Could not create singelton metric, error: ", err)
	}

	client, err := speed.NewPCPClient("simple")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	client.MustRegister(metric)

	client.MustStart()
	defer client.MustStop()

	fmt.Println("The metric is currently mapped as mmv.simple.simple.counter, to stop the mapping, press enter")
	_, _ = os.Stdin.Read(make([]byte, 1))
}
