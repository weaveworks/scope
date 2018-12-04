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
	"math/rand"
	"net/http"
	"time"

	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
)

func main() {
	r := promreporter.NewReporter(promreporter.Options{})

	// Note: `promreporter.DefaultSeparator` is "_".
	// Prometheus doesnt like metrics with "." or "-" in them.
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:         "my_service",
		Tags:           map[string]string{},
		CachedReporter: r,
		Separator:      promreporter.DefaultSeparator,
	}, 1*time.Second)
	defer closer.Close()

	counter := scope.Tagged(map[string]string{
		"foo": "bar",
	}).Counter("test_counter")

	gauge := scope.Tagged(map[string]string{
		"foo": "baz",
	}).Gauge("test_gauge")

	timer := scope.Tagged(map[string]string{
		"foo": "qux",
	}).Timer("test_timer_summary")

	histogram := scope.Tagged(map[string]string{
		"foo": "quk",
	}).Histogram("test_histogram", tally.DefaultBuckets)

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

	http.Handle("/metrics", r.HTTPHandler())
	fmt.Printf("Serving :8080/metrics\n")
	fmt.Printf("%v\n", http.ListenAndServe(":8080", nil))
	select {}
}
