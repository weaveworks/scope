package app_test

import (
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// StaticReport is used as a fixture in tests. It emulates an xfer.Collector.
type StaticReport struct{}

func (s StaticReport) Report() report.Report { return fixture.Report }
func (s StaticReport) Add(report.Report)     {}
func (s StaticReport) WaitOn(chan struct{})  {}
func (s StaticReport) UnWait(chan struct{})  {}
