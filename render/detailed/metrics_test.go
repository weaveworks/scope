package detailed_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

func TestNodeMetrics(t *testing.T) {
	inputs := []struct {
		name string
		node report.Node
		want []report.MetricRow
	}{
		{
			name: "process",
			node: fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
			want: []report.MetricRow{
				{
					ID:       process.CPUUsage,
					Label:    "CPU",
					Format:   "percent",
					Group:    "",
					Value:    0.01,
					Priority: 1,
					Metric:   &fixture.ClientProcess1CPUMetric,
				},
				{
					ID:       process.MemoryUsage,
					Label:    "Memory",
					Format:   "filesize",
					Group:    "",
					Value:    0.02,
					Priority: 2,
					Metric:   &fixture.ClientProcess1MemoryMetric,
				},
			},
		},
		{
			name: "container",
			node: fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
			want: []report.MetricRow{
				{
					ID:       docker.CPUTotalUsage,
					Label:    "CPU",
					Format:   "percent",
					Group:    "",
					Value:    0.03,
					Priority: 1,
					Metric:   &fixture.ClientContainerCPUMetric,
				},
				{
					ID:       docker.MemoryUsage,
					Label:    "Memory",
					Format:   "filesize",
					Group:    "",
					Value:    0.04,
					Priority: 2,
					Metric:   &fixture.ClientContainerMemoryMetric,
				},
			},
		},
		{
			name: "host",
			node: fixture.Report.Host.Nodes[fixture.ClientHostNodeID],
			want: []report.MetricRow{
				{
					ID:       host.CPUUsage,
					Label:    "CPU",
					Format:   "percent",
					Group:    "",
					Value:    0.07,
					Priority: 1,
					Metric:   &fixture.ClientHostCPUMetric,
				},
				{
					ID:       host.MemoryUsage,
					Label:    "Memory",
					Format:   "filesize",
					Group:    "",
					Value:    0.08,
					Priority: 2,
					Metric:   &fixture.ClientHostMemoryMetric,
				},
				{
					ID:       host.Load1,
					Label:    "Load (1m)",
					Group:    "load",
					Value:    0.09,
					Priority: 11,
					Metric:   &fixture.ClientHostLoad1Metric,
				},
			},
		},
		{
			name: "unknown topology",
			node: report.MakeNode(fixture.ClientContainerNodeID).WithTopology("foobar"),
			want: nil,
		},
	}
	for _, input := range inputs {
		have := detailed.NodeMetrics(fixture.Report, input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}

func TestMetricRowSummary(t *testing.T) {
	var (
		now    = time.Now()
		metric = report.MakeSingletonMetric(now, 1.234)
		row    = report.MetricRow{
			ID:       "id",
			Format:   "format",
			Group:    "group",
			Value:    1.234,
			Priority: 1,
			Metric:   &metric,
		}
		summary = row.Summary()
	)
	// summary should not have any samples
	if summary.Metric.Len() != 0 {
		t.Errorf("Expected summary to have no samples, but had %d", summary.Metric.Len())
	}
	// original metric should still have its samples
	if metric.Len() != 1 {
		t.Errorf("Expected original metric to still have it's samples, but had %d", metric.Len())
	}
	// summary should have all the same fields (minus the metric)
	summary.Metric = nil
	row.Metric = nil
	if !reflect.DeepEqual(summary, row) {
		t.Errorf("Expected summary to have same fields as original: %s", test.Diff(summary, row))
	}
}
