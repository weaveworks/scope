package app_test

import (
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"

	"golang.org/x/net/context"
)

// StaticReport is used as a fixture in tests. It emulates an xfer.Collector.
type StaticReport struct{}

func (s StaticReport) Report(context.Context) report.Report  { return fixture.Report }
func (s StaticReport) Add(context.Context, report.Report)    {}
func (s StaticReport) WaitOn(context.Context, chan struct{}) {}
func (s StaticReport) UnWait(context.Context, chan struct{}) {}
