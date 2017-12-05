package app

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportPath = flag.String("bench-report-path", "", "report file, or dir with files, to use for benchmarking (relative to this package)")
)

func disableProfiling() int {
	rate := runtime.MemProfileRate
	runtime.MemProfileRate = 0
	return rate
}

func enableProfiling(rate int) {
	runtime.MemProfileRate = rate
}

func readReportFiles(path string) ([]report.Report, error) {
	reports := []report.Report{}
	if err := filepath.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rpt, err := report.MakeFromFile(p)
			if err != nil {
				return err
			}
			reports = append(reports, rpt)
			return nil
		}); err != nil {
		return nil, err
	}
	return reports, nil
}

func BenchmarkReportUnmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()
		if _, err := readReportFiles(*benchReportPath); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReportMerge(b *testing.B) {
	p := disableProfiling()
	reports, err := readReportFiles(*benchReportPath)
	if err != nil {
		b.Fatal(err)
	}
	merger := NewSmartMerger()

	enableProfiling(p)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()
		merger.Merge(reports)
	}
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
	p := disableProfiling()
	r := fixture.Report
	if *benchReportPath != "" {
		reports, err := readReportFiles(*benchReportPath)
		if err != nil {
			b.Fatal(err)
		}
		r = NewSmartMerger().Merge(reports)
	}

	enableProfiling(p)
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

func BenchmarkTopologyControllers(b *testing.B) {
	benchmarkOneTopology(b, "kube-controllers")
}

func BenchmarkTopologyPods(b *testing.B) {
	benchmarkOneTopology(b, "pods")
}

func BenchmarkTopologyContainers(b *testing.B) {
	benchmarkOneTopology(b, "containers")
}

func BenchmarkTopologyProcesses(b *testing.B) {
	benchmarkOneTopology(b, "processes")
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
