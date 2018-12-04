// this example showcases speeds metric inference from strings property
package main

import (
	"flag"
	"log"
	"time"

	"github.com/performancecopilot/speed"
)

var timelimit = flag.Int("time", 60, "number of seconds to run for")

func main() {
	flag.Parse()

	c, err := speed.NewPCPClient("strings")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	m, err := c.RegisterString(
		"this.is.a.simple.counter.metric.to.demonstrate.the.RegisterString.function",
		10, speed.Int32Type, speed.CounterSemantics, speed.OneUnit)
	if err != nil {
		log.Fatal("Could not register string, error: ", err)
	}

	c.MustStart()
	defer c.MustStop()

	metric := m.(speed.SingletonMetric)
	for i := 0; i < *timelimit; i++ {
		val := metric.Val().(int32)
		val++
		metric.MustSet(val)
		time.Sleep(time.Second)
	}
}
