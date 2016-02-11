package detailed_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestMakeDetailedHostNode(t *testing.T) {
	renderableNode := render.HostRenderer.Render(fixture.Report)[render.MakeHostID(fixture.ClientHostID)]
	have := detailed.MakeNode(fixture.Report, renderableNode)

	containerImageNodeSummary, _ := detailed.MakeNodeSummary(
		render.ContainerImageRenderer.Render(fixture.Report)[render.MakeContainerImageID(fixture.ClientContainerImageName)].Node,
	)
	containerNodeSummary, _ := detailed.MakeNodeSummary(fixture.Report.Container.Nodes[fixture.ClientContainerNodeID])
	process1NodeSummary, _ := detailed.MakeNodeSummary(fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID])
	process1NodeSummary.Linkable = true
	process2NodeSummary, _ := detailed.MakeNodeSummary(fixture.Report.Process.Nodes[fixture.ClientProcess2NodeID])
	process2NodeSummary.Linkable = true
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:       render.MakeHostID(fixture.ClientHostID),
			Label:    "client",
			Linkable: true,
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
		Rank:     "hostname.com",
		Pseudo:   false,
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Containers",
				TopologyID: "containers",
				Columns:    []detailed.Column{docker.CPUTotalUsage, docker.MemoryUsage},
				Nodes:      []detailed.NodeSummary{containerNodeSummary},
			},
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns:    []detailed.Column{process.PID, process.CPUUsage, process.MemoryUsage},
				Nodes:      []detailed.NodeSummary{process1NodeSummary, process2NodeSummary},
			},
			{
				Label:      "Container Images",
				TopologyID: "containers-by-image",
				Columns:    []detailed.Column{render.ContainersKey},
				Nodes:      []detailed.NodeSummary{containerImageNodeSummary},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestMakeDetailedContainerNode(t *testing.T) {
	id := render.MakeContainerID(fixture.ServerContainerID)
	renderableNode, ok := render.ContainerRenderer.Render(fixture.Report)[id]
	if !ok {
		t.Fatalf("Node not found: %s", id)
	}
	have := detailed.MakeNode(fixture.Report, renderableNode)
	want := detailed.Node{
		NodeSummary: detailed.NodeSummary{
			ID:       id,
			Label:    "server",
			Linkable: true,
			Metadata: []detailed.MetadataRow{
				{ID: "docker_container_id", Value: fixture.ServerContainerID, Prime: true},
				{ID: "docker_container_state", Value: "running", Prime: true},
				{ID: "docker_image_id", Value: fixture.ServerContainerImageID},
			},
			DockerLabels: []detailed.MetadataRow{
				{ID: "label_" + render.AmazonECSContainerNameLabel, Value: `server`},
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
		Rank:     "imageid456",
		Pseudo:   false,
		Controls: []detailed.ControlInstance{},
		Children: []detailed.NodeSummaryGroup{
			{
				Label:      "Processes",
				TopologyID: "processes",
				Columns:    []detailed.Column{process.PID, process.CPUUsage, process.MemoryUsage},
				Nodes: []detailed.NodeSummary{
					{
						ID:       fmt.Sprintf("process:%s:%s", "server.hostname.com", fixture.ServerPID),
						Label:    "apache",
						Linkable: true,
						Metadata: []detailed.MetadataRow{
							{ID: process.PID, Value: fixture.ServerPID, Prime: true},
						},
						Metrics: []detailed.MetricRow{},
					},
				},
			},
		},
		Parents: []detailed.Parent{
			{
				ID:         render.MakeContainerImageID(fixture.ServerContainerImageName),
				Label:      fixture.ServerContainerImageName,
				TopologyID: "containers-by-image",
			},
			{
				ID:         render.MakeHostID(fixture.ServerHostName),
				Label:      fixture.ServerHostName,
				TopologyID: "hosts",
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
