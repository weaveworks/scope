package app

import (
	"flag"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportFile = flag.String("bench-report-file", "", "json report file to use for benchmarking (relative to this package)")
)

func loadReport() (report.Report, error) {
	if *benchReportFile == "" {
		return fixture.Report, nil
	}

	c, err := NewFileCollector(*benchReportFile, 0)
	if err != nil {
		return fixture.Report, err
	}

	return c.Report(context.Background())
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
