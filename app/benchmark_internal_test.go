package app

import (
	"flag"
	"net/http"
	"net/url"
	"testing"
	"time"

	"context"

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

	return c.Report(context.Background(), time.Now())
}

func BenchmarkTopologyList(b *testing.B) {
	benchmarkRender(b, func(report report.Report) {
		request := &http.Request{
			Form: url.Values{},
		}
		topologyRegistry.renderTopologies(report, request)
	})
}

func benchmarkRender(b *testing.B, f func(report.Report)) {
	report, err := loadReport()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		render.ResetCache()
		b.StartTimer()
		f(report)
	}
}

func BenchmarkTopologyHosts(b *testing.B) {
	benchmarkOneTopology(b, "hosts")
}

func BenchmarkTopologyContainers(b *testing.B) {
	benchmarkOneTopology(b, "containers")
}

func benchmarkOneTopology(b *testing.B, topologyID string) {
	benchmarkRender(b, func(report report.Report) {
		renderer, decorator, err := topologyRegistry.RendererForTopology(topologyID, url.Values{}, report)
		if err != nil {
			b.Fatal(err)
		}
		renderer.Render(report, decorator)
	})
}
