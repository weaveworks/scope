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
	renderableNode, _ := render.HostRenderer.Render(fixture.Report).Lookup(render.MakeHostID(fixture.ClientHostID))
	have := detailed.MakeNode(fixture.Report, renderableNode)

	containerImageNodeSummary, _ := detailed.MakeNodeSummary(fixture.Report.ContainerImage.Nodes[fixture.ClientContainerImageNodeID])
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
					Label: "Hostname",
					Value: "client.hostname.com",
				},
				{
					ID:    "os",
					Label: "Operating system",
					Value: "Linux",
				},
				{
					ID:    "local_networks",
					Label: "Local Networks",
					Value: "10.10.10.0/24",
				},
			},
			Metrics: []detailed.MetricRow{
				{
					ID:     host.CPUUsage,
					Format: "percent",
					Label:  "CPU",
					Value:  0.07,
					Metric: &fixture.ClientHostCPUMetric,
				},
				{
					ID:     host.MemoryUsage,
					Format: "filesize",
					Label:  "Memory",
					Value:  0.08,
					Metric: &fixture.ClientHostMemoryMetric,
				},
				{
					ID:     host.Load1,
					Group:  "load",
					Label:  "Load (1m)",
					Value:  0.09,
					Metric: &fixture.ClientHostLoad1Metric,
				},
				{
					ID:     host.Load5,
					Group:  "load",
					Label:  "Load (5m)",
					Value:  0.10,
					Metric: &fixture.ClientHostLoad5Metric,
				},
				{
					ID:     host.Load15,
					Label:  "Load (15m)",
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
				Label:      "Container Images",
				TopologyID: "containers-by-image",
				Columns:    []string{},
				Nodes:      []detailed.NodeSummary{containerImageNodeSummary},
			},
			{
				Label:      "Containers",
				TopologyID: "containers",
				Columns:    []string{docker.CPUTotalUsage, docker.MemoryUsage},
				Nodes:      []detailed.NodeSummary{containerNodeSummary},
			},
			{
				Label:      "Applications",
				TopologyID: "applications",
				Columns:    []string{process.PID, process.CPUUsage, process.MemoryUsage},
				Nodes:      []detailed.NodeSummary{process1NodeSummary, process2NodeSummary},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestMakeDetailedContainerNode(t *testing.T) {
	id := render.MakeContainerID(fixture.ServerContainerID)
	renderableNode, ok := render.ContainerRenderer.Render(fixture.Report).Lookup(id)
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
				{ID: "docker_container_id", Label: "ID", Value: fixture.ServerContainerID},
				{ID: "docker_image_id", Label: "Image ID", Value: fixture.ServerContainerImageID},
				{ID: "docker_container_state", Label: "State", Value: "running"},
				{ID: "label_" + render.AmazonECSContainerNameLabel, Label: fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), Value: `server`},
				{ID: "label_foo1", Label: `Label "foo1"`, Value: `bar1`},
				{ID: "label_foo2", Label: `Label "foo2"`, Value: `bar2`},
				{ID: "label_io.kubernetes.pod.name", Label: `Label "io.kubernetes.pod.name"`, Value: "ping/pong-b"},
			},
			Metrics: []detailed.MetricRow{
				{
					ID:     docker.CPUTotalUsage,
					Format: "percent",
					Label:  "CPU",
					Value:  0.05,
					Metric: &fixture.ServerContainerCPUMetric,
				},
				{
					ID:     docker.MemoryUsage,
					Format: "filesize",
					Label:  "Memory",
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
				Label:      "Applications",
				TopologyID: "applications",
				Columns:    []string{process.PID, process.CPUUsage, process.MemoryUsage},
				Nodes: []detailed.NodeSummary{
					{
						ID:       fmt.Sprintf("process:%s:%s", "server.hostname.com", fixture.ServerPID),
						Label:    "apache",
						Linkable: true,
						Metadata: []detailed.MetadataRow{
							{ID: process.PID, Label: "PID", Value: fixture.ServerPID},
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
