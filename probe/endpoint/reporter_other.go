// +build !linux

package endpoint

import "github.com/weaveworks/scope/report"

// Reporter dummy
type Reporter struct{}

// NewReporter makes a dummy
func NewReporter(conf ReporterConfig) *Reporter {
	return &Reporter{}
}

// Stop dummy
func (r *Reporter) Stop() {}

// Report implements Reporter.
func (r *Reporter) Report() (report.Report, error) {
	return report.MakeReport(), nil
}
