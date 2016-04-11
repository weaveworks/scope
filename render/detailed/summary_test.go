package detailed_test

import (
	"sort"
	"testing"
	"time"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestSummaries(t *testing.T) {
	{
		// Just a convenient source of some rendered nodes
		have := detailed.Summaries(fixture.Report, render.ProcessRenderer.Render(fixture.Report))
		// The ids of the processes rendered above
		expectedIDs := []string{
			fixture.ClientProcess1NodeID,
			fixture.ClientProcess2NodeID,
			fixture.ServerProcessNodeID,
			fixture.NonContainerProcessNodeID,
			expected.UnknownPseudoNode1ID,
			expected.UnknownPseudoNode2ID,
			render.IncomingInternetID,
			render.OutgoingInternetID,
		}
		sort.Strings(expectedIDs)

		// It should summarize each node
		ids := []string{}
		for id := range have {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		if !reflect.DeepEqual(expectedIDs, ids) {
			t.Fatalf("Expected Summaries to have summarized every node in the process renderer: %v, but got %v", expectedIDs, ids)
		}
	}

	// It should summarize nodes' metrics
	{
		t1, t2 := mtime.Now().Add(-1*time.Minute), mtime.Now()
		metric := report.MakeMetric().Add(t1, 1).Add(t2, 2)
		input := fixture.Report.Copy()

		input.Process.Nodes[fixture.ClientProcess1NodeID] = input.Process.Nodes[fixture.ClientProcess1NodeID].WithMetrics(report.Metrics{process.CPUUsage: metric})
		have := detailed.Summaries(input, render.ProcessRenderer.Render(input))

		node, ok := have[fixture.ClientProcess1NodeID]
		if !ok {
			t.Fatalf("Expected output to have the node we added the metric to")
		}

		var row report.MetricRow
		ok = false
		for _, metric := range node.Metrics {
			if metric.ID == process.CPUUsage {
				row = metric
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("Expected node to have the metric we added")
		}

		// Our summarized MetricRow
		want := report.MetricRow{
			ID:       process.CPUUsage,
			Label:    "CPU",
			Format:   "percent",
			Value:    2,
			Priority: 1,
			Metric: &report.Metric{
				Samples: nil,
				Min:     metric.Min,
				Max:     metric.Max,
				First:   metric.First,
				Last:    metric.Last,
			},
		}
		if !reflect.DeepEqual(want, row) {
			t.Fatalf("Expected to have summarized the node's metrics: %s", test.Diff(want, row))
		}
	}
}

func TestMakeNodeSummary(t *testing.T) {
	testcases := []struct {
		name  string
		input report.Node
		ok    bool
		want  detailed.NodeSummary
	}{
		{
			name:  "single process rendering",
			input: expected.RenderedProcesses[fixture.ClientProcess1NodeID],
			ok:    true,
			want: detailed.NodeSummary{
				ID:         fixture.ClientProcess1NodeID,
				Label:      fixture.Client1Name,
				LabelMinor: "client.hostname.com (10001)",
				Rank:       fixture.Client1Name,
				Shape:      "square",
				Metadata: []report.MetadataRow{
					{ID: process.PID, Label: "PID", Value: fixture.Client1PID, Priority: 1, Datatype: "number"},
				},
				Adjacency: report.MakeIDList(fixture.ServerProcessNodeID),
			},
		},
		{
			name:  "single container rendering",
			input: expected.RenderedContainers[fixture.ClientContainerNodeID],
			ok:    true,
			want: detailed.NodeSummary{
				ID:         fixture.ClientContainerNodeID,
				Label:      fixture.ClientContainerName,
				LabelMinor: fixture.ClientHostName,
				Rank:       fixture.ClientContainerImageName,
				Shape:      "hexagon",
				Linkable:   true,
				Metadata: []report.MetadataRow{
					{ID: docker.ContainerID, Label: "ID", Value: fixture.ClientContainerID, Priority: 1},
				},
				Adjacency: report.MakeIDList(fixture.ServerContainerNodeID),
			},
		},
		{
			name:  "single container image rendering",
			input: expected.RenderedContainerImages[fixture.ClientContainerImageNodeID],
			ok:    true,
			want: detailed.NodeSummary{
				ID:         fixture.ClientContainerImageNodeID,
				Label:      fixture.ClientContainerImageName,
				LabelMinor: "1 container",
				Rank:       fixture.ClientContainerImageName,
				Shape:      "hexagon",
				Linkable:   true,
				Stack:      true,
				Metadata: []report.MetadataRow{
					{ID: docker.ImageID, Label: "Image ID", Value: fixture.ClientContainerImageID, Priority: 1},
					{ID: report.Container, Label: "# Containers", Value: "1", Priority: 2, Datatype: "number"},
				},
				Adjacency: report.MakeIDList(fixture.ServerContainerImageNodeID),
			},
		},
		{
			name:  "single host rendering",
			input: expected.RenderedHosts[fixture.ClientHostNodeID],
			ok:    true,
			want: detailed.NodeSummary{
				ID:         fixture.ClientHostNodeID,
				Label:      "client",
				LabelMinor: "hostname.com",
				Rank:       "hostname.com",
				Shape:      "circle",
				Linkable:   true,
				Metadata: []report.MetadataRow{
					{ID: host.HostName, Label: "Hostname", Value: fixture.ClientHostName, Priority: 11},
				},
				Adjacency: report.MakeIDList(fixture.ServerHostNodeID),
			},
		},
	}
	for _, testcase := range testcases {
		have, ok := detailed.MakeNodeSummary(fixture.Report, testcase.input)
		if ok != testcase.ok {
			t.Errorf("%s: MakeNodeSummary failed: expected ok value to be: %v", testcase.name, testcase.ok)
			continue
		}

		if !reflect.DeepEqual(testcase.want, have) {
			t.Errorf("%s: Node Summary did not match: %s", testcase.name, test.Diff(testcase.want, have))
		}
	}
}
