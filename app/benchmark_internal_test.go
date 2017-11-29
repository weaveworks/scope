package app

import (
	"flag"
	"net/http"
	"net/url"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportFile = flag.String("bench-report-file", "", "report file to use for benchmarking (relative to this package)")
)

func BenchmarkTopologyList(b *testing.B) {
	benchmarkRender(b, func(report report.Report) {
		request := &http.Request{
			Form: url.Values{},
		}
		topologyRegistry.renderTopologies(report, request)
	})
}

func benchmarkRender(b *testing.B, f func(report.Report)) {
	r := fixture.Report
	if *benchReportFile != "" {
		var err error
		if r, err = report.MakeFromFile(*benchReportFile); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		render.ResetCache()
		b.StartTimer()
		f(r)
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
		renderer, filter, err := topologyRegistry.RendererForTopology(topologyID, url.Values{}, report)
		if err != nil {
			b.Fatal(err)
		}
		render.Render(report, renderer, filter)
	})
}
