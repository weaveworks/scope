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
	"log"
	"math/rand"
	"time"

	"github.com/uber-go/tally"
	statsdreporter "github.com/uber-go/tally/statsd"

	"github.com/cactus/go-statsd-client/statsd"
)

// To view statsd emitted metrics locally you can use
// netcat with "nc 8125 -l -u"
func main() {
	statter, err := statsd.NewBufferedClient("127.0.0.1:8125",
		"stats", 100*time.Millisecond, 1440)
	if err != nil {
		log.Fatalf("could not create statsd client: %v", err)
	}

	opts := statsdreporter.Options{}
	r := statsdreporter.NewReporter(statter, opts)

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:   "my-service",
		Tags:     map[string]string{},
		Reporter: r,
	}, 1*time.Second)
	defer closer.Close()

	counter := scope.Counter("test-counter")

	gauge := scope.Gauge("test-gauge")

	timer := scope.Timer("test-timer")

	histogram := scope.Histogram("test-histogram", tally.DefaultBuckets)

	go func() {
		for {
			counter.Inc(1)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			gauge.Update(rand.Float64() * 1000)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			tsw := timer.Start()
			hsw := histogram.Start()
			time.Sleep(time.Duration(rand.Float64() * float64(time.Second)))
			tsw.Stop()
			hsw.Stop()
		}
	}()

	select {}
}
