package detailed_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"

	"github.com/stretchr/testify/assert"
)

const (
	sampleMetricsGraphURL = "/prom/:orgID/notebook/new"
)

var (
	sampleUnknownNode = report.MakeNode("???").WithTopology("foo")
	samplePodNode     = report.MakeNode("noo").
				WithTopology(report.Pod).
				WithLatests(map[string]string{kubernetes.Namespace: "noospace"})
	sampleContainerNode = report.MakeNode("coo").
				WithTopology(report.Container).
				WithLatests(map[string]string{docker.ContainerName: "cooname"})
	sampleMetrics = []report.MetricRow{
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

func TestRenderMetricURLs_Pod(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, samplePodNode, sampleMetricsGraphURL)

	assert.Equal(t, 0, strings.Index(result.Metrics[0].URL, sampleMetricsGraphURL))
	// Double quotes are escaped since these are json marshaled strings
	contains := []string{"container_memory_usage_bytes", `pod_name=\"foo\"`, `namespace=\"noospace\"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[0].URL, url.QueryEscape(contain))
	}

	assert.Equal(t, 0, strings.Index(result.Metrics[1].URL, sampleMetricsGraphURL))
	contains = []string{"container_cpu_usage_seconds", `pod_name=\"foo\"`, `namespace=\"noospace\"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[1].URL, url.QueryEscape(contain))
	}
}

func TestRenderMetricURLs_Container(t *testing.T) {
	s := detailed.NodeSummary{Label: "foo", Metrics: sampleMetrics}
	result := detailed.RenderMetricURLs(s, sampleContainerNode, sampleMetricsGraphURL)

	assert.Equal(t, 0, strings.Index(result.Metrics[0].URL, sampleMetricsGraphURL))
	contains := []string{"container_memory_usage_bytes", `name=\"cooname\"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[0].URL, url.QueryEscape(contain))
	}

	assert.Equal(t, 0, strings.Index(result.Metrics[1].URL, sampleMetricsGraphURL))
	contains = []string{"container_cpu_usage_seconds", `name=\"cooname\"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[1].URL, url.QueryEscape(contain))
	}
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

	assert.Contains(t, result.Metrics[0].URL, "http://example.test/?q=")
	contains := []string{"container_memory_usage_bytes", `pod_name="foo"`, `namespace="noospace"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[0].URL, url.QueryEscape(contain))
	}
	assert.Contains(t, result.Metrics[1].URL, "http://example.test/?q=")
	contains = []string{"container_cpu_usage_seconds", `pod_name="foo"`, `namespace="noospace"`}
	for _, contain := range contains {
		assert.Contains(t, result.Metrics[1].URL, url.QueryEscape(contain))
	}
}
