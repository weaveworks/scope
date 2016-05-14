package app

import (
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/ugorji/go/codec"

	"$GITHUB_URI/render"
	"$GITHUB_URI/report"
	"$GITHUB_URI/test/fixture"
)

// StaticReport is used as a fixture in tests. It emulates an xfer.Collector.
type StaticReporter struct{ r report.Report }

func (s StaticReporter) Report() report.Report { return s.r }
func (s StaticReporter) WaitOn(chan struct{})  {}
func (s StaticReporter) UnWait(chan struct{})  {}

var (
	benchReportFile = flag.String("bench-report-file", "", "json report file to use for benchmarking (relative to this package)")
)

func loadReport() (report.Report, error) {
	if *benchReportFile == "" {
		return fixture.Report, nil
	}

	b, err := ioutil.ReadFile(*benchReportFile)
	if err != nil {
		return fixture.Report, err
	}
	rpt := report.MakeReport()
	decoder := codec.NewDecoderBytes(b, &codec.JsonHandle{})
	err = decoder.Decode(&rpt)
	return rpt, err
}

func BenchmarkTopologyList(b *testing.B) {
	report, err := loadReport()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	request := &http.Request{
		Form: url.Values{},
	}
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		render.ResetCache()
		b.StartTimer()
		topologyRegistry.renderTopologies(report, request)
	}
}
