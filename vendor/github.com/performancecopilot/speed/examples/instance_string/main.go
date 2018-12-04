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

	m, err := c.RegisterString("language[go, javascript, php].users", speed.Instances{
		"go":         1,
		"javascript": 100,
		"php":        10,
	}, speed.Uint64Type, speed.CounterSemantics, speed.OneUnit)
	if err != nil {
		log.Fatal("Could not register string, error: ", err)
	}

	c.MustStart()
	defer c.MustStop()

	metric := m.(speed.InstanceMetric)
	for i := 0; i < *timelimit; i++ {
		v, _ := metric.ValInstance("go")
		metric.MustSetInstance(v.(uint64)*2, "go")

		v, _ = metric.ValInstance("javascript")
		metric.MustSetInstance(v.(uint64)+10, "javascript")

		v, _ = metric.ValInstance("php")
		metric.MustSetInstance(v.(uint64)+1, "php")

		time.Sleep(time.Second)
	}
}
