package detailed_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestNodeMetrics(t *testing.T) {
	inputs := []struct {
		name string
		node report.Node
		want []detailed.MetricRow
	}{
		{
			name: "process",
			node: fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
			want: []detailed.MetricRow{
				{
					ID:     process.CPUUsage,
					Format: "percent",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.ClientProcess1CPUMetric,
				},
				{
					ID:     process.MemoryUsage,
					Format: "filesize",
					Group:  "",
					Value:  0.02,
					Metric: &fixture.ClientProcess1MemoryMetric,
				},
			},
		},
		{
			name: "container",
			node: fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
			want: []detailed.MetricRow{
				{
					ID:     docker.CPUTotalUsage,
					Format: "percent",
					Group:  "",
					Value:  0.03,
					Metric: &fixture.ClientContainerCPUMetric,
				},
				{
					ID:     docker.MemoryUsage,
					Format: "filesize",
					Group:  "",
					Value:  0.04,
					Metric: &fixture.ClientContainerMemoryMetric,
				},
			},
		},
		{
			name: "host",
			node: fixture.Report.Host.Nodes[fixture.ClientHostNodeID],
			want: []detailed.MetricRow{
				{
					ID:     host.CPUUsage,
					Format: "percent",
					Group:  "",
					Value:  0.07,
					Metric: &fixture.ClientHostCPUMetric,
				},
				{
					ID:     host.MemoryUsage,
					Format: "filesize",
					Group:  "",
					Value:  0.08,
					Metric: &fixture.ClientHostMemoryMetric,
				},
				{
					ID:     host.Load1,
					Group:  "load",
					Value:  0.09,
					Metric: &fixture.ClientHostLoad1Metric,
				},
			},
		},
		{
			name: "unknown topology",
			node: report.MakeNode().WithTopology("foobar").WithID(fixture.ClientContainerNodeID),
			want: nil,
		},
	}
	for _, input := range inputs {
		have := detailed.NodeMetrics(input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}

func TestMetricRowSummary(t *testing.T) {
	var (
		now    = time.Now()
		metric = report.MakeMetric().Add(now, 1.234)
		row    = detailed.MetricRow{
			ID:     "id",
			Format: "format",
			Group:  "group",
			Value:  1.234,
			Metric: &metric,
		}
		summary = row.Summary()
	)
	// summary should have all the same fields
	if row.ID != summary.ID || row.Format != summary.Format || row.Group != summary.Group || row.Value != summary.Value {
		t.Errorf("Expected summary to have same fields as original: %#v, but had %#v", row, summary)
	}
	// summary should not have any samples
	if summary.Metric.Len() != 0 {
		t.Errorf("Expected summary to have no samples, but had %d", summary.Metric.Len())
	}
	// original metric should still have its samples
	if metric.Len() != 1 {
		t.Errorf("Expected original metric to still have it's samples, but had %d", metric.Len())
	}
}
