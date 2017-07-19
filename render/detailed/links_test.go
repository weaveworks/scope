package detailed_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"

	"github.com/stretchr/testify/assert"
)

const (
	sampleMetricsGraphURL = "/prom/:orgID/notebook/new"
)

var (
	sampleUnknownNode = report.MakeNode("???").WithTopology("foo")
	samplePodNode     = report.MakeNode("noo").WithTopology(report.Pod)
	sampleMetrics     = []report.MetricRow{
		{ID: docker.MemoryUsage},
		{ID: docker.CPUTotalUsage},
	}
)

func TestRenderMetricURLs_Disabled(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, samplePodNode, "")

	assert.Empty(t, result.Metrics[0].URL)
	assert.Empty(t, result.Metrics[1].URL)
}

func TestRenderMetricURLs_UnknownTopology(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, sampleUnknownNode, sampleMetricsGraphURL)

	assert.Empty(t, result.Metrics[0].URL)
	assert.Empty(t, result.Metrics[1].URL)
}

func TestRenderMetricURLs(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, samplePodNode, sampleMetricsGraphURL)

	u := "/prom/:orgID/notebook/new/%7B%22cells%22%3A%5B%7B%22queries%22%3A%5B%22sum%28container_memory_usage_bytes%7Bpod_name%3D%5C%22foo%5C%22%7D%29%22%5D%7D%5D%7D"
	assert.Equal(t, u, result.Metrics[0].URL)
	u = "/prom/:orgID/notebook/new/%7B%22cells%22%3A%5B%7B%22queries%22%3A%5B%22sum%28rate%28container_cpu_usage_seconds_total%7Bpod_name%3D%5C%22foo%5C%22%7D%5B1m%5D%29%29%22%5D%7D%5D%7D"
	assert.Equal(t, u, result.Metrics[1].URL)
}

func TestRenderMetricURLs_EmptyMetrics(t *testing.T) {
	result := detailed.RenderMetricURLs(detailed.NodeSummary{}, samplePodNode, sampleMetricsGraphURL)

	m := result.Metrics[0]
	assert.Equal(t, docker.CPUTotalUsage, m.ID)
	assert.Equal(t, "CPU", m.Label)
	assert.NotEmpty(t, m.URL)
	assert.True(t, m.ValueEmpty)
	assert.Equal(t, float64(1), m.Priority)

	m = result.Metrics[1]
	assert.NotEmpty(t, m.URL)
	assert.True(t, m.ValueEmpty)
	assert.Equal(t, float64(2), m.Priority)
}

func TestRenderMetricURLs_CombinedEmptyMetrics(t *testing.T) {
	s := detailed.NodeSummary{
		Label:   "foo",
		Metrics: []report.MetricRow{{ID: docker.MemoryUsage, Priority: 1}},
	}
	result := detailed.RenderMetricURLs(s, samplePodNode, sampleMetricsGraphURL)

	assert.NotEmpty(t, result.Metrics[0].URL)
	assert.False(t, result.Metrics[0].ValueEmpty)

	assert.NotEmpty(t, result.Metrics[1].URL)
	assert.True(t, result.Metrics[1].ValueEmpty)
	assert.Equal(t, float64(2), result.Metrics[1].Priority) // first empty metric starts at non-empty prio + 1
}

func TestRenderMetricURLs_QueryReplacement(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, samplePodNode, "http://example.test/?q=:query")

	u := "http://example.test/?q=sum%28container_memory_usage_bytes%7Bpod_name%3D%22foo%22%7D%29"
	assert.Equal(t, u, result.Metrics[0].URL)
	u = "http://example.test/?q=sum%28rate%28container_cpu_usage_seconds_total%7Bpod_name%3D%22foo%22%7D%5B1m%5D%29%29"
	assert.Equal(t, u, result.Metrics[1].URL)
}
