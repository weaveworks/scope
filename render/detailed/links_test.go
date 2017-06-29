package detailed_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"

	"github.com/stretchr/testify/assert"
)

var (
	sampleReport      = report.Report{}
	samplePodNode     = report.MakeNode("noo").WithTopology(report.Pod)
	sampleUnknownNode = report.MakeNode("???").WithTopology("foo")
)

func TestNodeMetricLinks_DefaultDisabled(t *testing.T) {
	links := detailed.NodeMetricLinks(sampleReport, samplePodNode)
	assert.Nil(t, links)
}

func TestNodeMetricLinks_UnknownTopology(t *testing.T) {
	detailed.SetMetricsGraphURL("/foo")

	links := detailed.NodeMetricLinks(sampleReport, sampleUnknownNode)
	assert.Nil(t, links)
}

func TestNodeMetricLinks(t *testing.T) {
	detailed.SetMetricsGraphURL("/foo")
	defer detailed.SetMetricsGraphURL("")
	expected := []detailed.MetricLink{
		{ID: docker.CPUTotalUsage, Label: "CPU", Priority: 1, URL: ""},
		{ID: docker.MemoryUsage, Label: "Memory", Priority: 2, URL: ""},
	}

	links := detailed.NodeMetricLinks(sampleReport, samplePodNode)
	assert.Equal(t, expected, links)
}

func TestRenderMetricLinks_UnknownTopology(t *testing.T) {
	summary := detailed.NodeSummary{}

	result := detailed.RenderMetricLinks(summary, sampleUnknownNode)
	assert.Equal(t, summary, result)
}

func TestRenderMetricLinks_Pod(t *testing.T) {
	detailed.SetMetricsGraphURL("/prom/:orgID/notebook/new")
	defer detailed.SetMetricsGraphURL("")
	summary := detailed.NodeSummary{Label: "woo", MetricLinks: detailed.NodeMetricLinks(sampleReport, samplePodNode)}

	result := detailed.RenderMetricLinks(summary, samplePodNode)
	assert.Equal(t,
		"/prom/:orgID/notebook/new/%7B%22cells%22:%5B%7B%22queries%22:%5B%22sum%28rate%28container_cpu_usage_seconds_total%7Bpod_name=%5C%22woo%5C%22%7D%5B1m%5D%29%29%22%5D%7D%5D%7D",
		result.MetricLinks[0].URL)
	assert.Equal(t,
		"/prom/:orgID/notebook/new/%7B%22cells%22:%5B%7B%22queries%22:%5B%22sum%28container_memory_usage_bytes%7Bpod_name=%5C%22woo%5C%22%7D%29%22%5D%7D%5D%7D",
		result.MetricLinks[1].URL)
}

func TestRenderMetricLinks_QueryReplacement(t *testing.T) {
	detailed.SetMetricsGraphURL("/foo/:orgID/bar?q=:query")
	defer detailed.SetMetricsGraphURL("")
	summary := detailed.NodeSummary{Label: "boo", MetricLinks: detailed.NodeMetricLinks(sampleReport, samplePodNode)}

	result := detailed.RenderMetricLinks(summary, samplePodNode)
	assert.Equal(t,
		"/foo/:orgID/bar?q=sum%28rate%28container_cpu_usage_seconds_total%7Bpod_name=%22boo%22%7D%5B1m%5D%29%29",
		result.MetricLinks[0].URL)
	assert.Equal(t,
		"/foo/:orgID/bar?q=sum%28container_memory_usage_bytes%7Bpod_name=%22boo%22%7D%29",
		result.MetricLinks[1].URL)
}
