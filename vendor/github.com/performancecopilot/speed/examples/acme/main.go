// A golang implementation of the acme factory examples from python and Java APIs
//
// The python implementation is in mmv.py in PCP core
// (https://github.com/performancecopilot/pcp/blob/master/src/python/pcp/mmv.py#L21-L70)
//
// The Java implementation is in examples in parfait core
// (https://github.com/performancecopilot/parfait/tree/master/examples/acme)
//
// To run the python version of the example that exits do
// go run examples/acme/main.go
//
// To run the java version of the example that runs forever, simply add a --forever
// flag
// go run examples/acme/main.go --forever
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/performancecopilot/speed"
)

var runforever bool

func init() {
	flag.BoolVar(&runforever, "forever", false, "if enabled, runs the forever running version of this example")
}

func main() {
	flag.Parse()

	if runforever {
		forever()
	} else {
		serial()
	}
}

func serial() {
	instances := []string{"Anvils", "Rockets", "Giant_Rubber_Bands"}

	indom, err := speed.NewPCPInstanceDomain(
		"Acme Products",
		instances,
		"Acme products",
		"Most popular products produced by the Acme Corporation",
	)
	if err != nil {
		log.Fatal("Could not create indom, error: ", err)
	}

	countmetric, err := speed.NewPCPInstanceMetric(
		speed.Instances{
			"Anvils":             0,
			"Rockets":            0,
			"Giant_Rubber_Bands": 0,
		},
		"products.count",
		indom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.OneUnit,
		"Acme factory product throughput",
		`Monotonic increasing counter of products produced in the Acme Corporation
		 factory since starting the Acme production application.  Quality guaranteed.`,
	)
	if err != nil {
		log.Fatal("Could not create countmetric, error: ", err)
	}

	timemetric, err := speed.NewPCPInstanceMetric(
		speed.Instances{
			"Anvils":             0,
			"Rockets":            0,
			"Giant_Rubber_Bands": 0,
		},
		"products.time",
		indom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.MicrosecondUnit,
		"Machine time spent producing Acme products",
	)
	if err != nil {
		log.Fatal("Could not create timemetric, error: ", err)
	}

	client, err := speed.NewPCPClient("acme")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	client.MustRegisterIndom(indom)
	client.MustRegister(countmetric)
	client.MustRegister(timemetric)

	client.MustStart()
	defer client.MustStop()

	time.Sleep(time.Second * 5)
	err = countmetric.SetInstance(42, "Anvils")
	if err != nil {
		log.Fatal("Could not set countmetric[\"Anvils\"], error: ", err)
	}
	time.Sleep(time.Second * 5)
}

// ProductBuilder is based on ProductBuilder in the parfait example
// https://github.com/performancecopilot/parfait/blob/master/examples/acme/src/main/java/ProductBuilder.java
type ProductBuilder struct {
	completed speed.Counter
	totalTime speed.Gauge
	bound     int
	name      string
}

// NewProductBuilder creates a new instance of ProductBuilder
func NewProductBuilder(name string, client speed.Client) *ProductBuilder {
	completed, err := speed.NewPCPCounter(0, "products."+name+".count")
	if err != nil {
		log.Fatal("Could not create completed, error: ", err)
	}

	totalTime, err := speed.NewPCPGauge(0, "products."+name+".time")
	if err != nil {
		log.Fatal("Could not create totalTime, error: ", err)
	}

	client.MustRegister(completed)
	client.MustRegister(totalTime)

	return &ProductBuilder{
		name:      name,
		bound:     500,
		completed: completed,
		totalTime: totalTime,
	}
}

// Difficulty sets the upper bound on the sleep time
func (p *ProductBuilder) Difficulty(bound int) {
	p.bound = bound
}

// Build sleeps for a random time, then adds that value to totalTime
func (p *ProductBuilder) Build() {
	elapsed := rand.Intn(p.bound)

	time.Sleep(time.Duration(elapsed) * time.Millisecond)

	p.totalTime.MustInc(float64(elapsed))
	p.completed.Up()
}

// Start starts an infinite loop calling Build and logging the value of completed
func (p *ProductBuilder) Start() {
	for {
		p.Build()
		fmt.Printf("Built %d %s\n", p.completed.Val(), p.name)
	}
}

func forever() {
	client, err := speed.NewPCPClient("acme")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	rockets := NewProductBuilder("Rockets", client)
	anvils := NewProductBuilder("Anvils", client)
	gbrs := NewProductBuilder("Giant_Rubber_Bands", client)

	rockets.Difficulty(4500)
	anvils.Difficulty(1500)
	gbrs.Difficulty(2500)

	go func() {
		rockets.Start()
	}()

	go func() {
		anvils.Start()
	}()

	go func() {
		gbrs.Start()
	}()

	client.MustStart()
	defer client.MustStop()

	// block forever
	// TODO: maybe use signal.Notify and shut down gracefully
	select {}
}
