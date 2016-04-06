package detailed_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func child(t *testing.T, r render.Renderer, id string) detailed.NodeSummary {
	s, ok := detailed.MakeNodeSummary(r.Render(fixture.Report)[id])
	if !ok {
		t.Fatalf("Expected node %s to be summarizable, but wasn't", id)
	}
	return s.SummarizeMetrics()
}

func TestMakeDetailedHostNode(t *testing.T) {
	renderableNodes := render.HostRenderer.Render(fixture.Report)
	renderableNode := renderableNodes[fixture.ClientHostNodeID]
	have := detailed.MakeNode("hosts", fixture.Report, renderableNodes, renderableNode)

	containerImageNodeSummary := child(t, render.ContainerImageRenderer, fixture.ClientContainerImageNodeID)
	containerNodeSummary := child(t, render.ContainerRenderer, fixture.ClientContainerNodeID)
	process1NodeSummary := child(t, render.ProcessRenderer, fixture.ClientProcess1NodeID)
	process1NodeSummary.Linkable = true
	process2NodeSummary := child(t, render.ProcessRenderer, fixture.ClientProcess2NodeID)
	process2NodeSummary.Linkable = true
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:         fixture.ClientHostNodeID,
			Label:      "client",
			LabelMinor: "hostname.com",
			Rank:       "hostname.com",
			Pseudo:     false,
			Shape:      "circle",
			Linkable:   true,
			Adjacency:  report.MakeIDList(fixture.ServerHostNodeID),
			Metadata: []detailed.MetadataRow{
				{
					ID:    "host_name",
					Value: "client.hostname.com",
				},
				{
					ID:    "os",
					Value: "Linux",
				},
				{
					ID:    "local_networks",
					Value: "10.10.10.0/24",
				},
			},
			Metrics: []detailed.MetricRow{
				{
					ID:     host.CPUUsage,
					Format: "percent",
					Value:  0.07,
					Metric: &fixture.ClientHostCPUMetric,
				},
				{
					ID:     host.MemoryUsage,
					Format: "filesize",
					Value:  0.08,
					Metric: &fixture.ClientHostMemoryMetric,
				},
				{
					ID:     host.Load1,
					Group:  "load",
					Value:  0.09,
					Metric: &fixture.ClientHostLoad1Metric,
				},
				{
					ID:     host.Load5,
					Group:  "load",
					Value:  0.10,
					Metric: &fixture.ClientHostLoad5Metric,
				},
				{
					ID:     host.Load15,
					Group:  "load",
					Value:  0.11,
					Metric: &fixture.ClientHostLoad15Metric,
				},
			},
		},
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Containers",
				TopologyID: "containers",
				Columns:    []detailed.Column{detailed.MakeColumn(docker.CPUTotalUsage), detailed.MakeColumn(docker.MemoryUsage)},
				Nodes:      []detailed.NodeSummary{containerNodeSummary},
			},
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns:    []detailed.Column{detailed.MakeColumn(process.PID), detailed.MakeColumn(process.CPUUsage), detailed.MakeColumn(process.MemoryUsage)},
				Nodes:      []detailed.NodeSummary{process1NodeSummary, process2NodeSummary},
			},
			{
				Label:      "Container Images",
				TopologyID: "containers-by-image",
				Columns: []detailed.Column{
					{ID: report.Container, Label: detailed.Label(report.Container), DefaultSort: true},
				},
				Nodes: []detailed.NodeSummary{containerImageNodeSummary},
			},
		},
		Connections: []detailed.NodeSummaryGroup{
			{
				ID:         "incoming-connections",
				TopologyID: "hosts",
				Label:      "Inbound",
				Columns:    detailed.NormalColumns,
				Nodes:      []detailed.NodeSummary{},
			},
			{
				ID:         "outgoing-connections",
				TopologyID: "hosts",
				Label:      "Outbound",
				Columns:    detailed.NormalColumns,
				Nodes: []detailed.NodeSummary{
					{
						ID:         fixture.ServerHostNodeID,
						Label:      "server",
						LabelMinor: "hostname.com",
						Rank:       "hostname.com",
						Shape:      "circle",
						Linkable:   true,
						Adjacency:  report.MakeIDList(render.OutgoingInternetID),
						Metadata: []detailed.MetadataRow{
							{
								ID:       "port",
								Value:    "80",
								Datatype: "number",
							},
							{
								ID:       "count",
								Value:    "2",
								Datatype: "number",
							},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestMakeDetailedContainerNode(t *testing.T) {
	id := fixture.ServerContainerNodeID
	renderableNodes := render.ContainerRenderer.Render(fixture.Report)
	renderableNode, ok := renderableNodes[id]
	if !ok {
		t.Fatalf("Node not found: %s", id)
	}
	have := detailed.MakeNode("containers", fixture.Report, renderableNodes, renderableNode)

	serverProcessNodeSummary := child(t, render.ProcessRenderer, fixture.ServerProcessNodeID)
	serverProcessNodeSummary.Linkable = true
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:         id,
			Label:      "server",
			LabelMinor: "server.hostname.com",
			Shape:      "hexagon",
			Linkable:   true,
			Pseudo:     false,
			Metadata: []detailed.MetadataRow{
				{ID: "docker_container_id", Value: fixture.ServerContainerID, Prime: true},
				{ID: "docker_container_state_human", Value: "running", Prime: true},
				{ID: "docker_image_id", Value: fixture.ServerContainerImageID},
			},
			DockerLabels: []detailed.MetadataRow{
				{ID: "label_" + detailed.AmazonECSContainerNameLabel, Value: `server`},
				{ID: "label_foo1", Value: `bar1`},
				{ID: "label_foo2", Value: `bar2`},
				{ID: "label_io.kubernetes.pod.name", Value: "ping/pong-b"},
			},
			Metrics: []detailed.MetricRow{
				{
					ID:     docker.CPUTotalUsage,
					Format: "percent",
					Value:  0.05,
					Metric: &fixture.ServerContainerCPUMetric,
				},
				{
					ID:     docker.MemoryUsage,
					Format: "filesize",
					Value:  0.06,
					Metric: &fixture.ServerContainerMemoryMetric,
				},
			},
		},
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns:    []detailed.Column{detailed.MakeColumn(process.PID), detailed.MakeColumn(process.CPUUsage), detailed.MakeColumn(process.MemoryUsage)},
				Nodes:      []detailed.NodeSummary{serverProcessNodeSummary},
			},
		},
		Parents: []detailed.Parent{
			{
				ID:         fixture.ServerContainerImageNodeID,
				Label:      fixture.ServerContainerImageName,
				TopologyID: "containers-by-image",
			},
			{
				ID:         fixture.ServerHostNodeID,
				Label:      fixture.ServerHostName,
				TopologyID: "hosts",
			},
		},
		Connections: []detailed.NodeSummaryGroup{
			{
				ID:         "incoming-connections",
				TopologyID: "containers",
				Label:      "Inbound",
				Columns:    detailed.NormalColumns,
				Nodes: []detailed.NodeSummary{
					{
						ID:         fixture.ClientContainerNodeID,
						Label:      "client",
						LabelMinor: "client.hostname.com",
						Shape:      "hexagon",
						Linkable:   true,
						Adjacency:  report.MakeIDList(fixture.ServerContainerNodeID),
						Metadata: []detailed.MetadataRow{
							{
								ID:       "port",
								Value:    "80",
								Datatype: "number",
							},
							{
								ID:       "count",
								Value:    "2",
								Datatype: "number",
							},
						},
					},
					{
						ID:         render.IncomingInternetID,
						Label:      render.InboundMajor,
						LabelMinor: render.InboundMinor,
						Rank:       render.IncomingInternetID,
						Shape:      "cloud",
						Linkable:   true,
						Pseudo:     true,
						Adjacency:  report.MakeIDList(fixture.ServerContainerNodeID),
						Metadata: []detailed.MetadataRow{
							{
								ID:       "port",
								Value:    "80",
								Datatype: "number",
							},
							{
								ID:       "count",
								Value:    "1",
								Datatype: "number",
							},
						},
					},
				},
			},
			{
				ID:         "outgoing-connections",
				TopologyID: "containers",
				Label:      "Outbound",
				Columns:    detailed.NormalColumns,
				Nodes:      []detailed.NodeSummary{},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
