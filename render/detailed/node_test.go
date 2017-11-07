package detailed_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func child(t *testing.T, r render.Renderer, id string) detailed.NodeSummary {
	s, ok := detailed.MakeNodeSummary(report.RenderContext{Report: fixture.Report}, r.Render(fixture.Report, nil).Nodes[id])
	if !ok {
		t.Fatalf("Expected node %s to be summarizable, but wasn't", id)
	}
	return s.SummarizeMetrics()
}

func connectionID(nodeID string, addr string) string {
	return fmt.Sprintf("%s-%s-%s-%d", nodeID, addr, "", 80)
}

func TestMakeDetailedHostNode(t *testing.T) {
	renderableNodes := render.HostRenderer.Render(fixture.Report, nil).Nodes
	renderableNode := renderableNodes[fixture.ClientHostNodeID]
	have := detailed.MakeNode("hosts", report.RenderContext{Report: fixture.Report}, renderableNodes, renderableNode)

	containerImageNodeSummary := child(t, render.ContainerImageRenderer, expected.ClientContainerImageNodeID)
	containerNodeSummary := child(t, render.ContainerRenderer, fixture.ClientContainerNodeID)
	process1NodeSummary := child(t, render.ProcessRenderer, fixture.ClientProcess1NodeID)
	process1NodeSummary.Linkable = true
	process2NodeSummary := child(t, render.ProcessRenderer, fixture.ClientProcess2NodeID)
	process2NodeSummary.Linkable = true
	podNodeSummary := child(t, render.PodRenderer, fixture.ClientPodNodeID)
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
			Metadata: []report.MetadataRow{
				{
					ID:       "host_name",
					Label:    "Hostname",
					Value:    "client.hostname.com",
					Priority: 11,
				},
				{
					ID:       "os",
					Label:    "OS",
					Value:    "Linux",
					Priority: 12,
				},
				{
					ID:       "local_networks",
					Label:    "Local Networks",
					Value:    "10.10.10.0/24",
					Priority: 13,
				},
			},
			Metrics: []report.MetricRow{
				{
					ID:       host.CPUUsage,
					Label:    "CPU",
					Format:   "percent",
					Value:    0.07,
					Priority: 1,
					Metric:   &fixture.ClientHostCPUMetric,
				},
				{
					ID:       host.MemoryUsage,
					Label:    "Memory",
					Format:   "filesize",
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
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Pods",
				TopologyID: "pods",
				Columns: []detailed.Column{
					{ID: kubernetes.State, Label: "State"},
					{ID: report.Container, Label: "# Containers", Datatype: report.Number},
					{ID: kubernetes.IP, Label: "IP", Datatype: report.IP},
				},
				Nodes: []detailed.NodeSummary{podNodeSummary},
			},
			{
				Label:      "Containers",
				TopologyID: "containers",
				Columns: []detailed.Column{
					{ID: docker.CPUTotalUsage, Label: "CPU", Datatype: report.Number},
					{ID: docker.MemoryUsage, Label: "Memory", Datatype: report.Number},
				},
				Nodes: []detailed.NodeSummary{containerNodeSummary},
			},
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns: []detailed.Column{
					{ID: process.PID, Label: "PID", Datatype: report.Number},
					{ID: process.CPUUsage, Label: "CPU", Datatype: report.Number},
					{ID: process.MemoryUsage, Label: "Memory", Datatype: report.Number},
				},
				Nodes: []detailed.NodeSummary{process1NodeSummary, process2NodeSummary},
			},
			{
				Label:      "Container Images",
				TopologyID: "containers-by-image",
				Columns:    []detailed.Column{},
				Nodes:      []detailed.NodeSummary{containerImageNodeSummary},
			},
		},
		Connections: []detailed.ConnectionsSummary{
			{
				ID:          "incoming-connections",
				TopologyID:  "hosts",
				Label:       "Inbound",
				Columns:     detailed.NormalColumns,
				Connections: []detailed.Connection{},
			},
			{
				ID:         "outgoing-connections",
				TopologyID: "hosts",
				Label:      "Outbound",
				Columns:    detailed.NormalColumns,
				Connections: []detailed.Connection{
					{
						ID:         connectionID(fixture.ServerHostNodeID, ""),
						NodeID:     fixture.ServerHostNodeID,
						Label:      "server",
						LabelMinor: "hostname.com",
						Linkable:   true,
						Metadata: []report.MetadataRow{
							{
								ID:    "port",
								Value: "80",
							},
							{
								ID:    "count",
								Value: "2",
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
	renderableNodes := render.ContainerWithImageNameRenderer.Render(fixture.Report, nil).Nodes
	renderableNode, ok := renderableNodes[id]
	if !ok {
		t.Fatalf("Node not found: %s", id)
	}
	have := detailed.MakeNode("containers", report.RenderContext{Report: fixture.Report}, renderableNodes, renderableNode)

	serverProcessNodeSummary := child(t, render.ProcessRenderer, fixture.ServerProcessNodeID)
	serverProcessNodeSummary.Linkable = true
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:         id,
			Label:      "server",
			LabelMinor: "server.hostname.com",
			Rank:       fixture.ServerContainerImageName,
			Shape:      "hexagon",
			Linkable:   true,
			Pseudo:     false,
			Metadata: []report.MetadataRow{
				{ID: "docker_image_name", Label: "Image", Value: fixture.ServerContainerImageName, Priority: 1},
				{ID: "docker_container_state_human", Label: "State", Value: "running", Priority: 3},
				{ID: "docker_container_id", Label: "ID", Value: fixture.ServerContainerID, Priority: 10, Truncate: 12},
			},
			Metrics: []report.MetricRow{
				{
					ID:       docker.CPUTotalUsage,
					Label:    "CPU",
					Format:   "percent",
					Value:    0.05,
					Priority: 1,
					Metric:   &fixture.ServerContainerCPUMetric,
				},
				{
					ID:       docker.MemoryUsage,
					Label:    "Memory",
					Format:   "filesize",
					Value:    0.06,
					Priority: 2,
					Metric:   &fixture.ServerContainerMemoryMetric,
				},
			},
			Parents: []detailed.Parent{
				{
					ID:         expected.ServerContainerImageNodeID,
					Label:      fixture.ServerContainerImageName,
					TopologyID: "containers-by-image",
				},
				{
					ID:         fixture.ServerHostNodeID,
					Label:      fixture.ServerHostName,
					TopologyID: "hosts",
				},
				{
					ID:         fixture.ServerPodNodeID,
					Label:      "pong-b",
					TopologyID: "pods",
				},
			},
		},
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns: []detailed.Column{
					{ID: process.PID, Label: "PID", Datatype: report.Number},
					{ID: process.CPUUsage, Label: "CPU", Datatype: report.Number},
					{ID: process.MemoryUsage, Label: "Memory", Datatype: report.Number},
				},
				Nodes: []detailed.NodeSummary{serverProcessNodeSummary},
			},
		},
		Connections: []detailed.ConnectionsSummary{
			{
				ID:         "incoming-connections",
				TopologyID: "containers",
				Label:      "Inbound",
				Columns:    detailed.NormalColumns,
				Connections: []detailed.Connection{
					{
						ID:         connectionID(fixture.ClientContainerNodeID, ""),
						NodeID:     fixture.ClientContainerNodeID,
						Label:      "client",
						LabelMinor: "client.hostname.com",
						Linkable:   true,
						Metadata: []report.MetadataRow{
							{
								ID:    "port",
								Value: "80",
							},
							{
								ID:    "count",
								Value: "2",
							},
						},
					},
					{
						ID:       connectionID(render.IncomingInternetID, fixture.RandomClientIP),
						NodeID:   render.IncomingInternetID,
						Label:    fixture.RandomClientIP,
						Linkable: true,
						Metadata: []report.MetadataRow{
							{
								ID:    "port",
								Value: "80",
							},
							{
								ID:    "count",
								Value: "1",
							},
						},
					},
				},
			},
			{
				ID:          "outgoing-connections",
				TopologyID:  "containers",
				Label:       "Outbound",
				Columns:     detailed.NormalColumns,
				Connections: []detailed.Connection{},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestMakeDetailedPodNode(t *testing.T) {
	id := fixture.ServerPodNodeID
	renderableNodes := render.PodRenderer.Render(fixture.Report, nil).Nodes
	renderableNode, ok := renderableNodes[id]
	if !ok {
		t.Fatalf("Node not found: %s", id)
	}
	have := detailed.MakeNode("pods", report.RenderContext{Report: fixture.Report}, renderableNodes, renderableNode)

	containerNodeSummary := child(t, render.ContainerWithImageNameRenderer, fixture.ServerContainerNodeID)
	serverProcessNodeSummary := child(t, render.ProcessRenderer, fixture.ServerProcessNodeID)
	serverProcessNodeSummary.Linkable = true // Temporary workaround for: https://github.com/weaveworks/scope/issues/1295
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:         id,
			Label:      "pong-b",
			LabelMinor: "1 container",
			Rank:       "ping/pong-b",
			Shape:      "heptagon",
			Linkable:   true,
			Pseudo:     false,
			Metadata: []report.MetadataRow{
				{ID: "kubernetes_state", Label: "State", Value: "running", Priority: 2},
				{ID: "container", Label: "# Containers", Value: "1", Priority: 4, Datatype: report.Number},
				{ID: "kubernetes_namespace", Label: "Namespace", Value: "ping", Priority: 5},
			},
			Parents: []detailed.Parent{
				{
					ID:         fixture.ServerHostNodeID,
					Label:      fixture.ServerHostName,
					TopologyID: "hosts",
				},
				{
					ID:         fixture.ServiceNodeID,
					Label:      fixture.ServiceName,
					TopologyID: "services",
				},
			},
		},
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Containers",
				TopologyID: "containers",
				Columns: []detailed.Column{
					{ID: docker.CPUTotalUsage, Label: "CPU", Datatype: report.Number},
					{ID: docker.MemoryUsage, Label: "Memory", Datatype: report.Number},
				},
				Nodes: []detailed.NodeSummary{containerNodeSummary},
			},
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns: []detailed.Column{
					{ID: process.PID, Label: "PID", Datatype: report.Number},
					{ID: process.CPUUsage, Label: "CPU", Datatype: report.Number},
					{ID: process.MemoryUsage, Label: "Memory", Datatype: report.Number},
				},
				Nodes: []detailed.NodeSummary{serverProcessNodeSummary},
			},
		},
		Connections: []detailed.ConnectionsSummary{
			{
				ID:         "incoming-connections",
				TopologyID: "pods",
				Label:      "Inbound",
				Columns:    detailed.NormalColumns,
				Connections: []detailed.Connection{
					{
						ID:         connectionID(fixture.ClientPodNodeID, ""),
						NodeID:     fixture.ClientPodNodeID,
						Label:      "pong-a",
						LabelMinor: "1 container",
						Linkable:   true,
						Metadata: []report.MetadataRow{
							{
								ID:    "port",
								Value: "80",
							},
							{
								ID:    "count",
								Value: "2",
							},
						},
					},
					{
						ID:       connectionID(render.IncomingInternetID, fixture.RandomClientIP),
						NodeID:   render.IncomingInternetID,
						Label:    fixture.RandomClientIP,
						Linkable: true,
						Metadata: []report.MetadataRow{
							{
								ID:    "port",
								Value: "80",
							},
							{
								ID:    "count",
								Value: "1",
							},
						},
					},
				},
			},
			{
				ID:          "outgoing-connections",
				TopologyID:  "pods",
				Label:       "Outbound",
				Columns:     detailed.NormalColumns,
				Connections: []detailed.Connection{},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
