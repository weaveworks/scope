// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"
	"time"

	"github.com/uber-go/tally"
)

type printStatsReporter struct{}

func newPrintStatsReporter() tally.StatsReporter {
	return &printStatsReporter{}
}

func (r *printStatsReporter) ReportCounter(name string, _ map[string]string, value int64) {
	fmt.Printf("count %s %d\n", name, value)
}

func (r *printStatsReporter) ReportGauge(name string, _ map[string]string, value float64) {
	fmt.Printf("gauge %s %f\n", name, value)
}

func (r *printStatsReporter) ReportTimer(name string, _ map[string]string, interval time.Duration) {
	fmt.Printf("timer %s %s\n", name, interval.String())
}

func (r *printStatsReporter) ReportHistogramValueSamples(
	name string,
	_ map[string]string,
	_ tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	fmt.Printf("histogram %s bucket lower %f upper %f samples %d\n",
		name, bucketLowerBound, bucketUpperBound, samples)
}

func (r *printStatsReporter) ReportHistogramDurationSamples(
	name string,
	_ map[string]string,
	_ tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	fmt.Printf("histogram %s bucket lower %v upper %v samples %d\n",
		name, bucketLowerBound, bucketUpperBound, samples)
}

func (r *printStatsReporter) Capabilities() tally.Capabilities {
	return r
}

func (r *printStatsReporter) Reporting() bool {
	return true
}

func (r *printStatsReporter) Tagging() bool {
	return false
}

func (r *printStatsReporter) Flush() {
	fmt.Printf("flush\n")
}

func main() {
	reporter := newPrintStatsReporter()
	rootScope, closer := tally.NewRootScope(tally.ScopeOptions{
		Reporter: reporter,
	}, time.Second)
	defer closer.Close()
	subScope := rootScope.SubScope("requests")

	bighand := time.NewTicker(time.Millisecond * 2300)
	littlehand := time.NewTicker(time.Millisecond * 10)
	hugehand := time.NewTicker(time.Millisecond * 5100)

	measureThing := rootScope.Gauge("thing")
	timings := rootScope.Timer("timings")
	tickCounter := subScope.Counter("ticks")

	// Spin forever, watch report get called
	go func() {
		for {
			select {
			case <-bighand.C:
				measureThing.Update(42.1)
			case <-littlehand.C:
				tickCounter.Inc(1)
			case <-hugehand.C:
				timings.Record(3200 * time.Millisecond)
			}
		}
	}()

	select {}
}
