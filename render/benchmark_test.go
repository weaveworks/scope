package render_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportFile       = flag.String("bench-report-file", "", "json report file to use for benchmarking (relative to this package)")
	benchmarkRenderResult map[string]render.RenderableNode
	benchmarkStatsResult  render.Stats
)

func BenchmarkEndpointRender(b *testing.B) { benchmarkRender(b, render.EndpointRenderer) }
func BenchmarkEndpointStats(b *testing.B)  { benchmarkStats(b, render.EndpointRenderer) }
func BenchmarkProcessRender(b *testing.B)  { benchmarkRender(b, render.ProcessRenderer) }
func BenchmarkProcessStats(b *testing.B)   { benchmarkStats(b, render.ProcessRenderer) }
func BenchmarkProcessWithContainerNameRender(b *testing.B) {
	benchmarkRender(b, render.ProcessWithContainerNameRenderer)
}
func BenchmarkProcessWithContainerNameStats(b *testing.B) {
	benchmarkStats(b, render.ProcessWithContainerNameRenderer)
}
func BenchmarkProcessNameRender(b *testing.B) { benchmarkRender(b, render.ProcessNameRenderer) }
func BenchmarkProcessNameStats(b *testing.B)  { benchmarkStats(b, render.ProcessNameRenderer) }
func BenchmarkContainerRender(b *testing.B)   { benchmarkRender(b, render.ContainerRenderer) }
func BenchmarkContainerStats(b *testing.B)    { benchmarkStats(b, render.ContainerRenderer) }
func BenchmarkContainerWithImageNameRender(b *testing.B) {
	benchmarkRender(b, render.ContainerWithImageNameRenderer)
}
func BenchmarkContainerWithImageNameStats(b *testing.B) {
	benchmarkStats(b, render.ContainerWithImageNameRenderer)
}
func BenchmarkContainerImageRender(b *testing.B) { benchmarkRender(b, render.ContainerImageRenderer) }
func BenchmarkContainerImageStats(b *testing.B)  { benchmarkStats(b, render.ContainerImageRenderer) }
func BenchmarkContainerHostnameRender(b *testing.B) {
	benchmarkRender(b, render.ContainerHostnameRenderer)
}
func BenchmarkContainerHostnameStats(b *testing.B) {
	benchmarkStats(b, render.ContainerHostnameRenderer)
}
func BenchmarkHostRender(b *testing.B)       { benchmarkRender(b, render.HostRenderer) }
func BenchmarkHostStats(b *testing.B)        { benchmarkStats(b, render.HostRenderer) }
func BenchmarkPodRender(b *testing.B)        { benchmarkRender(b, render.PodRenderer) }
func BenchmarkPodStats(b *testing.B)         { benchmarkStats(b, render.PodRenderer) }
func BenchmarkPodServiceRender(b *testing.B) { benchmarkRender(b, render.PodServiceRenderer) }
func BenchmarkPodServiceStats(b *testing.B)  { benchmarkStats(b, render.PodServiceRenderer) }

func benchmarkRender(b *testing.B, r render.Renderer) {
	report, err := loadReport()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkRenderResult = r.Render(report)
		if len(benchmarkRenderResult) == 0 {
			b.Errorf("Rendered topology contained no nodes")
		}
	}
}

func benchmarkStats(b *testing.B, r render.Renderer) {
	report, err := loadReport()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// No way to tell if this was successful :(
		benchmarkStatsResult = r.Stats(report)
	}
}

func loadReport() (report.Report, error) {
	if *benchReportFile == "" {
		return fixture.Report, nil
	}

	var rpt report.Report
	b, err := ioutil.ReadFile(*benchReportFile)
	if err != nil {
		return rpt, err
	}
	err = json.Unmarshal(b, &rpt)
	return rpt, err
}
