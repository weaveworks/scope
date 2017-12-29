package app

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportPath = flag.String("bench-report-path", "", "report file, or dir with files, to use for benchmarking (relative to this package)")
)

func readReportFiles(b *testing.B, path string) []report.Report {
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
		b.Fatal(err)
	}
	return reports
}

func BenchmarkReportUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readReportFiles(b, *benchReportPath)
	}
}

func upgradeReports(reports []report.Report) []report.Report {
	upgraded := make([]report.Report, len(reports))
	for i, r := range reports {
		upgraded[i] = r.Upgrade()
	}
	return upgraded
}

func BenchmarkReportUpgrade(b *testing.B) {
	reports := readReportFiles(b, *benchReportPath)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		upgradeReports(reports)
	}
}

func BenchmarkReportMerge(b *testing.B) {
	reports := upgradeReports(readReportFiles(b, *benchReportPath))
	merger := NewSmartMerger()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger.Merge(reports)
	}
}

func getReport(b *testing.B) report.Report {
	r := fixture.Report
	if *benchReportPath != "" {
		r = NewSmartMerger().Merge(upgradeReports(readReportFiles(b, *benchReportPath)))
	}
	return r
}

func benchmarkRender(b *testing.B, f func(report.Report)) {
	r := getReport(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		render.ResetCache()
		b.StartTimer()
		f(r)
	}
}

func renderForTopology(b *testing.B, topologyID string, report report.Report) report.Nodes {
	renderer, filter, err := topologyRegistry.RendererForTopology(topologyID, url.Values{}, report)
	if err != nil {
		b.Fatal(err)
	}
	return render.Render(report, renderer, filter).Nodes
}

func benchmarkRenderTopology(b *testing.B, topologyID string) {
	benchmarkRender(b, func(report report.Report) {
		renderForTopology(b, topologyID, report)
	})
}

func BenchmarkRenderList(b *testing.B) {
	benchmarkRender(b, func(report report.Report) {
		topologyRegistry.renderTopologies(report, &http.Request{Form: url.Values{}})
	})
}

func BenchmarkRenderHosts(b *testing.B) {
	benchmarkRenderTopology(b, "hosts")
}

func BenchmarkRenderControllers(b *testing.B) {
	benchmarkRenderTopology(b, "kube-controllers")
}

func BenchmarkRenderPods(b *testing.B) {
	benchmarkRenderTopology(b, "pods")
}

func BenchmarkRenderContainers(b *testing.B) {
	benchmarkRenderTopology(b, "containers")
}

func BenchmarkRenderProcesses(b *testing.B) {
	benchmarkRenderTopology(b, "processes")
}

func BenchmarkRenderProcessNames(b *testing.B) {
	benchmarkRenderTopology(b, "processes-by-name")
}

func benchmarkSummarizeTopology(b *testing.B, topologyID string) {
	r := getReport(b)
	rc := detailed.RenderContext{Report: r}
	nodes := renderForTopology(b, topologyID, r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detailed.Summaries(rc, nodes)
	}
}

func BenchmarkSummarizeHosts(b *testing.B) {
	benchmarkSummarizeTopology(b, "hosts")
}

func BenchmarkSummarizeControllers(b *testing.B) {
	benchmarkSummarizeTopology(b, "kube-controllers")
}

func BenchmarkSummarizePods(b *testing.B) {
	benchmarkSummarizeTopology(b, "pods")
}

func BenchmarkSummarizeContainers(b *testing.B) {
	benchmarkSummarizeTopology(b, "containers")
}

func BenchmarkSummarizeProcesses(b *testing.B) {
	benchmarkSummarizeTopology(b, "processes")
}

func BenchmarkSummarizeProcessNames(b *testing.B) {
	benchmarkSummarizeTopology(b, "processes-by-name")
}
