package main

import (
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

// StaticReport is used as know test data in api tests.
type StaticReport struct{}

func (s StaticReport) Report() report.Report {
	return test.Report
}
