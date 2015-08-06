package main

import (
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

// StaticReport is used as a fixture in tests. It emulates an xfer.Collector.
type StaticReport struct{}

func (s StaticReport) Report() report.Report { return test.Report }

func (s StaticReport) Add(report.Report) {}
