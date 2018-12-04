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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/uber-go/tally"
	"github.com/uber-go/tally/m3"
	m3thrift "github.com/uber-go/tally/m3/thrift"

	validator "gopkg.in/validator.v2"
	yaml "gopkg.in/yaml.v2"
)

var configFileArg = flag.String("config", "config.yaml", "YAML config file path")

type config struct {
	M3 m3.Configuration `yaml:"m3"`
}

func configFromFile(file string) (*config, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	cfg := &config{}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func main() {
	flag.Parse()
	configFile := *configFileArg
	if configFile == "" {
		flag.Usage()
		return
	}

	cfg, err := configFromFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file %s: %v", configFile, err)
	}

	r, err := cfg.M3.NewReporter()
	if err != nil {
		log.Fatalf("failed to create reporter: %v", err)
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		CachedReporter: r,
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

	srv, err := newLocalM3Server("127.0.0.1:6396", m3.Compact, func(b *m3thrift.MetricBatch) {
		for _, m := range b.Metrics {
			tags := make(map[string]string)
			for tag := range b.CommonTags {
				tags[tag.GetTagName()] = tag.GetTagValue()
			}
			for tag := range m.Tags {
				tags[tag.GetTagName()] = tag.GetTagValue()
			}
			metVal := m.GetMetricValue()
			switch {
			case metVal != nil && metVal.Count != nil:
				fmt.Printf("counter value: %d, tags: %v\n", metVal.Count.GetI64Value(), tags)
			case metVal != nil && metVal.Gauge != nil && metVal.Gauge.I64Value != nil:
				fmt.Printf("gauge value: %d, tags: %v\n", metVal.Gauge.GetI64Value(), tags)
			case metVal != nil && metVal.Gauge != nil && metVal.Gauge.DValue != nil:
				fmt.Printf("gauge value: %f, tags: %v\n", metVal.Gauge.GetDValue(), tags)
			case metVal != nil && metVal.Timer != nil:
				fmt.Printf("timer value: %v, tags: %v\n", time.Duration(metVal.Timer.GetI64Value()), tags)
			}
		}
	})
	if err != nil {
		log.Fatalf("failed to create test listen server: %v", err)
	}
	if err := srv.Serve(); err != nil {
		log.Fatalf("failed to serve test listen server: %v", err)
	}
}
