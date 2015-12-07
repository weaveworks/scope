package detailed_test

import (
	"reflect"
	"testing"

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
					Label:  "CPU Usage",
					Format: "percent",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.CPUMetric,
				},
				{
					ID:     process.MemoryUsage,
					Label:  "Memory Usage",
					Format: "filesize",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.MemoryMetric,
				},
			},
		},
		{
			name: "container",
			node: fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
			want: []detailed.MetricRow{
				{
					ID:     docker.CPUTotalUsage,
					Label:  "CPU Usage",
					Format: "percent",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.CPUMetric,
				},
				{
					ID:     docker.MemoryUsage,
					Label:  "Memory Usage",
					Format: "filesize",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.MemoryMetric,
				},
			},
		},
		{
			name: "host",
			node: fixture.Report.Host.Nodes[fixture.ClientHostNodeID],
			want: []detailed.MetricRow{
				{
					ID:     host.CPUUsage,
					Label:  "CPU Usage",
					Format: "percent",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.CPUMetric,
				},
				{
					ID:     host.MemUsage,
					Label:  "Memory Usage",
					Format: "filesize",
					Group:  "",
					Value:  0.01,
					Metric: &fixture.MemoryMetric,
				},
				{
					ID:     host.Load1,
					Label:  "Load (1m)",
					Group:  "load",
					Value:  0.01,
					Metric: &fixture.LoadMetric,
				},
				{
					ID:     host.Load5,
					Label:  "Load (5m)",
					Group:  "load",
					Value:  0.01,
					Metric: &fixture.LoadMetric,
				},
				{
					ID:     host.Load15,
					Label:  "Load (15m)",
					Group:  "load",
					Value:  0.01,
					Metric: &fixture.LoadMetric,
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
