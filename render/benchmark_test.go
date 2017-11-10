package render_test

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportFile       = flag.String("bench-report-file", "", "json report file to use for benchmarking (relative to this package)")
	benchmarkRenderResult render.Nodes
)

func BenchmarkEndpointRender(b *testing.B) { benchmarkRender(b, render.EndpointRenderer) }
func BenchmarkProcessRender(b *testing.B)  { benchmarkRender(b, render.ProcessRenderer) }
func BenchmarkProcessWithContainerNameRender(b *testing.B) {
	benchmarkRender(b, render.ProcessWithContainerNameRenderer)
}
func BenchmarkProcessNameRender(b *testing.B) { benchmarkRender(b, render.ProcessNameRenderer) }
func BenchmarkContainerRender(b *testing.B)   { benchmarkRender(b, render.ContainerRenderer) }
func BenchmarkContainerWithImageNameRender(b *testing.B) {
	benchmarkRender(b, render.ContainerWithImageNameRenderer)
}
func BenchmarkContainerImageRender(b *testing.B) {
	benchmarkRender(b, render.ContainerImageRenderer)
}
func BenchmarkContainerHostnameRender(b *testing.B) {
	benchmarkRender(b, render.ContainerHostnameRenderer)
}
func BenchmarkHostRender(b *testing.B) { benchmarkRender(b, render.HostRenderer) }
func BenchmarkPodRender(b *testing.B)  { benchmarkRender(b, render.PodRenderer) }
func BenchmarkPodServiceRender(b *testing.B) {
	benchmarkRender(b, render.PodServiceRenderer)
}

func benchmarkRender(b *testing.B, r render.Renderer) {

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
		benchmarkRenderResult = r.Render(report, FilterNoop)
		if len(benchmarkRenderResult.Nodes) == 0 {
			b.Errorf("Rendered topology contained no nodes")
		}
	}
}

func loadReport() (report.Report, error) {
	if *benchReportFile == "" {
		return fixture.Report, nil
	}

	b, err := ioutil.ReadFile(*benchReportFile)
	if err != nil {
		return rpt, err
	}
	rpt := report.MakeReport()
	err = codec.NewDecoderBytes(b, &codec.JsonHandle{}).Decode(&rpt)
	return rpt, err
}
