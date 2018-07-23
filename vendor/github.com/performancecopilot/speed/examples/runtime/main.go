package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/performancecopilot/speed"
)

// refresh interval
const interval = time.Millisecond

// list of memory-related metric domains
var memMetricInstances = []string{
	"Alloc",
	"TotalAlloc",
	"Sys",
	"Lookups",
	"Mallocs",
	"Frees",
	"HeapAlloc",
	"HeapSys",
	"HeapIdle",
	"HeapInuse",
	"HeapReleased",
	"HeapObjects",
	"StackInuse",
	"StackSys",
	"MSpanInuse",
	"MSpanSys",
	"MCacheInuse",
	"MCacheSys",
	"BuckHashSys",
	"GCSys",
	"OtherSys",
	"NextGC",
	"LastGC",
	"PauseTotalNs",
	"PauseNs",
	"PauseEnd",
	"NumGC",
}

func main() {
	cpuIndom, err := speed.NewPCPInstanceDomain(
		"CPU Metrics",
		[]string{"CGoCalls", "Goroutines"},
	)
	if err != nil {
		log.Fatal("Could not create cpuIndom, error: ", err)
	}

	cpuMetric, err := speed.NewPCPInstanceMetric(
		speed.Instances{
			"CGoCalls":   0,
			"Goroutines": 0,
		},
		"cpu",
		cpuIndom,
		speed.Int64Type,
		speed.CounterSemantics,
		speed.OneUnit,
	)
	if err != nil {
		log.Fatal("Could not create cpuMetric, error: ", err)
	}

	memIndom, err := speed.NewPCPInstanceDomain("Memory Metrics", memMetricInstances)
	if err != nil {
		log.Fatal("Could not create memIndom, error: ", err)
	}

	memInsts := speed.Instances{}
	for _, v := range memMetricInstances {
		memInsts[v] = 0
	}
	memMetric, err := speed.NewPCPInstanceMetric(
		memInsts,
		"mem",
		memIndom,
		speed.Uint64Type,
		speed.CounterSemantics,
		speed.OneUnit,
	)
	if err != nil {
		log.Fatal("Could not create memMetric, error: ", err)
	}

	client, err := speed.NewPCPClient("runtime")
	if err != nil {
		log.Fatal("Could not create client, error: ", err)
	}

	client.MustRegister(cpuMetric)
	client.MustRegister(memMetric)
	client.MustStart()
	defer client.MustStop()

	mStats := runtime.MemStats{}

	c := time.Tick(interval)
	go func() {
		for range c {
			cpuMetric.MustSetInstance(runtime.NumCgoCall(), "CGoCalls")
			cpuMetric.MustSetInstance(runtime.NumGoroutine(), "Goroutines")

			runtime.ReadMemStats(&mStats)
			memMetric.MustSetInstance(mStats.Alloc, "Alloc")
			memMetric.MustSetInstance(mStats.TotalAlloc, "TotalAlloc")
			memMetric.MustSetInstance(mStats.Sys, "Sys")
			memMetric.MustSetInstance(mStats.Mallocs, "Mallocs")
			memMetric.MustSetInstance(mStats.Frees, "Frees")
			memMetric.MustSetInstance(mStats.HeapAlloc, "HeapAlloc")
			memMetric.MustSetInstance(mStats.HeapSys, "HeapSys")
			memMetric.MustSetInstance(mStats.HeapIdle, "HeapIdle")
			memMetric.MustSetInstance(mStats.HeapInuse, "HeapInuse")
			memMetric.MustSetInstance(mStats.HeapReleased, "HeapReleased")
			memMetric.MustSetInstance(mStats.HeapObjects, "HeapObjects")
			memMetric.MustSetInstance(mStats.StackInuse, "StackInuse")
			memMetric.MustSetInstance(mStats.StackSys, "StackSys")
			memMetric.MustSetInstance(mStats.MSpanInuse, "MSpanInuse")
			memMetric.MustSetInstance(mStats.MSpanSys, "MSpanSys")
			memMetric.MustSetInstance(mStats.MCacheInuse, "MCacheInuse")
			memMetric.MustSetInstance(mStats.MCacheSys, "MCacheSys")
			memMetric.MustSetInstance(mStats.BuckHashSys, "BuckHashSys")
			memMetric.MustSetInstance(mStats.GCSys, "GCSys")
			memMetric.MustSetInstance(mStats.OtherSys, "OtherSys")
			memMetric.MustSetInstance(mStats.NextGC, "NextGC")
			memMetric.MustSetInstance(mStats.LastGC, "LastGC")
			memMetric.MustSetInstance(mStats.PauseTotalNs, "PauseTotalNs")
			memMetric.MustSetInstance(mStats.PauseNs[(mStats.NumGC+255)%256], "PauseNs")
			memMetric.MustSetInstance(mStats.PauseEnd[(mStats.NumGC+255)%256], "PauseEnd")
			memMetric.MustSetInstance(uint64(mStats.NumGC), "NumGC")
		}
	}()

	fmt.Println("To stop the mapping, press enter")
	_, _ = os.Stdin.Read(make([]byte, 1))
}
