package detailed_test

//
//import (
//	"context"
//	"sort"
//	"testing"
//	"time"
//
//	"github.com/weaveworks/common/mtime"
//	"github.com/weaveworks/common/test"
//	"github.com/weaveworks/scope/probe/docker"
//	"github.com/weaveworks/scope/probe/host"
//	"github.com/weaveworks/scope/probe/process"
//	"github.com/weaveworks/scope/render"
//	"github.com/weaveworks/scope/render/detailed"
//	"github.com/weaveworks/scope/render/expected"
//	"github.com/weaveworks/scope/report"
//	"github.com/weaveworks/scope/test/fixture"
//	"github.com/weaveworks/scope/test/reflect"
//)
//
//func TestSummaries(t *testing.T) {
//	{
//		// Just a convenient source of some rendered nodes
//		have := detailed.Summaries(context.Background(), detailed.RenderContext{Report: fixture.Report}, render.ProcessRenderer.Render(context.Background(), fixture.Report).Nodes)
//		// The ids of the processes rendered above
//		expectedIDs := []string{
//			fixture.ClientProcess1NodeID,
//			fixture.ClientProcess2NodeID,
//			fixture.ServerProcessNodeID,
//			fixture.NonContainerProcessNodeID,
//			render.IncomingInternetID,
//			render.OutgoingInternetID,
//		}
//		sort.Strings(expectedIDs)
//
//		// It should summarize each node
//		ids := []string{}
//		for id := range have {
//			ids = append(ids, id)
//		}
//		sort.Strings(ids)
//		if !reflect.DeepEqual(expectedIDs, ids) {
//			t.Fatalf("Expected Summaries to have summarized every node in the process renderer: %v, but got %v", expectedIDs, ids)
//		}
//	}
//
//	// It should summarize nodes' metrics
//	{
//		t1, t2 := mtime.Now().Add(-1*time.Minute), mtime.Now()
//		metric := report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 1}, {Timestamp: t2, Value: 2}})
//		input := fixture.Report.Copy()
//		processNode := input.Process.Nodes[fixture.ClientProcess1NodeID]
//		processNode.Metrics = processNode.Metrics.Copy()
//		processNode.Metrics[process.CPUUsage] = metric
//		input.Process.Nodes[fixture.ClientProcess1NodeID] = processNode
//		have := detailed.Summaries(context.Background(), detailed.RenderContext{Report: input}, render.ProcessRenderer.Render(context.Background(), input).Nodes)
//
//		node, ok := have[fixture.ClientProcess1NodeID]
//		if !ok {
//			t.Fatalf("Expected output to have the node we added the metric to")
//		}
//
//		var row report.MetricRow
//		ok = false
//		for _, metric := range node.Metrics {
//			if metric.ID == process.CPUUsage {
//				row = metric
//				ok = true
//				break
//			}
//		}
//		if !ok {
//			t.Fatalf("Expected node to have the metric we added")
//		}
//
//		// Our summarized MetricRow
//		want := report.MetricRow{
//			ID:       process.CPUUsage,
//			Label:    "CPU",
//			Format:   "percent",
//			Value:    2,
//			Priority: 1,
//			Metric: &report.Metric{
//				Samples: nil,
//				Min:     metric.Min,
//				Max:     metric.Max,
//			},
//		}
//		if !reflect.DeepEqual(want, row) {
//			t.Fatalf("Expected to have summarized the node's metrics: %s", test.Diff(want, row))
//		}
//	}
//}
//
//func TestMakeNodeSummary(t *testing.T) {
//	testcases := []struct {
//		name  string
//		input report.Node
//		ok    bool
//		want  detailed.NodeSummary
//	}{
//		{
//			name:  "single process rendering",
//			input: expected.RenderedProcesses[fixture.ClientProcess1NodeID],
//			ok:    true,
//			want: detailed.NodeSummary{
//				BasicNodeSummary: detailed.BasicNodeSummary{
//					ID:         fixture.ClientProcess1NodeID,
//					Label:      fixture.Client1Name,
//					LabelMinor: "client.hostname.com (10001)",
//					Rank:       fixture.Client1Name,
//					Shape:      "square",
//					Tag:        "",
//				},
//				Metadata: []report.MetadataRow{
//					{ID: process.PID, Label: "PID", Value: fixture.Client1PID, Priority: 1, Datatype: report.Number},
//				},
//				Adjacency: report.MakeIDList(fixture.ServerProcessNodeID),
//			},
//		},
//		{
//			name:  "single container rendering",
//			input: expected.RenderedContainers[fixture.ClientContainerNodeID],
//			ok:    true,
//			want: detailed.NodeSummary{
//				BasicNodeSummary: detailed.BasicNodeSummary{
//					ID:         fixture.ClientContainerNodeID,
//					Label:      fixture.ClientContainerName,
//					LabelMinor: fixture.ClientHostName,
//					Rank:       fixture.ClientContainerImageName,
//					Shape:      "hexagon",
//					Tag:        "",
//				},
//				Metadata: []report.MetadataRow{
//					{ID: docker.ImageName, Label: "Image name", Value: fixture.ClientContainerImageName, Priority: 2},
//					{ID: docker.ContainerID, Label: "ID", Value: fixture.ClientContainerID, Priority: 11, Truncate: 12},
//				},
//				Adjacency: report.MakeIDList(fixture.ServerContainerNodeID),
//			},
//		},
//		{
//			name:  "single container image rendering",
//			input: expected.RenderedContainerImages[expected.ClientContainerImageNodeID],
//			ok:    true,
//			want: detailed.NodeSummary{
//				BasicNodeSummary: detailed.BasicNodeSummary{
//					ID:         expected.ClientContainerImageNodeID,
//					Label:      fixture.ClientContainerImageName,
//					LabelMinor: "1 container",
//					Rank:       fixture.ClientContainerImageName,
//					Shape:      "hexagon",
//					Tag:        "",
//					Stack:      true,
//				},
//				Metadata: []report.MetadataRow{
//					{ID: report.Container, Label: "# Containers", Value: "1", Priority: 2, Datatype: report.Number},
//				},
//				Adjacency: report.MakeIDList(expected.ServerContainerImageNodeID),
//			},
//		},
//		{
//			name:  "single host rendering",
//			input: expected.RenderedHosts[fixture.ClientHostNodeID],
//			ok:    true,
//			want: detailed.NodeSummary{
//				BasicNodeSummary: detailed.BasicNodeSummary{
//					ID:         fixture.ClientHostNodeID,
//					Label:      "client",
//					LabelMinor: "hostname.com",
//					Rank:       "hostname.com",
//					Shape:      "circle",
//					Tag:        "",
//				},
//				Metadata: []report.MetadataRow{
//					{ID: host.HostName, Label: "Hostname", Value: fixture.ClientHostName, Priority: 11},
//				},
//				Adjacency: report.MakeIDList(fixture.ServerHostNodeID),
//			},
//		},
//		{
//			name:  "group node rendering",
//			input: expected.RenderedProcessNames[fixture.ServerName],
//			ok:    true,
//			want: detailed.NodeSummary{
//				BasicNodeSummary: detailed.BasicNodeSummary{
//					ID:         "apache",
//					Label:      "apache",
//					LabelMinor: "1 process",
//					Rank:       "apache",
//					Shape:      "square",
//					Tag:        "",
//					Stack:      true,
//				},
//			},
//		},
//	}
//	for _, testcase := range testcases {
//		have, ok := detailed.MakeNodeSummary(detailed.RenderContext{Report: fixture.Report}, testcase.input)
//		if ok != testcase.ok {
//			t.Errorf("%s: MakeNodeSummary failed: expected ok value to be: %v", testcase.name, testcase.ok)
//			continue
//		}
//
//		if !reflect.DeepEqual(testcase.want, have) {
//			t.Errorf("%s: Node Summary did not match: %s", testcase.name, test.Diff(testcase.want, have))
//		}
//	}
//}
//
//func TestMakeNodeSummaryNoMetadata(t *testing.T) {
//	processNameTopology := render.MakeGroupNodeTopology(report.Process, process.Name)
//	for topology, id := range map[string]string{
//		render.Pseudo:         render.MakePseudoNodeID("id"),
//		report.Process:        report.MakeProcessNodeID("ip-123-45-6-100", "1234"),
//		report.Container:      report.MakeContainerNodeID("0001accbecc2c95e650fe641926fb923b7cc307a71101a1200af3759227b6d7d"),
//		report.ContainerImage: report.MakeContainerImageNodeID("0001accbecc2c95e650fe641926fb923b7cc307a71101a1200af3759227b6d7d"),
//		report.Pod:            report.MakePodNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.Service:        report.MakeServiceNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.Deployment:     report.MakeDeploymentNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.DaemonSet:      report.MakeDaemonSetNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.StatefulSet:    report.MakeStatefulSetNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.CronJob:        report.MakeCronJobNodeID("005e2999-d429-11e7-8535-0a41257e78e8"),
//		report.ECSTask:        report.MakeECSTaskNodeID("arn:aws:ecs:us-east-1:012345678910:task/1dc5c17a-422b-4dc4-b493-371970c6c4d6"),
//		report.ECSService:     report.MakeECSServiceNodeID("cluster", "service"),
//		report.SwarmService:   report.MakeSwarmServiceNodeID("0001accbecc2c95e650fe641926fb923b7cc307a71101a1200af3759227b6d7d"),
//		report.Host:           report.MakeHostNodeID("ip-123-45-6-100"),
//		report.Overlay:        report.MakeOverlayNodeID("", "3e:ca:14:ca:12:5c"),
//		processNameTopology:   "/home/weave/scope",
//	} {
//		summary, b := detailed.MakeNodeSummary(detailed.RenderContext{}, report.MakeNode(id).WithTopology(topology))
//		switch {
//		case !b:
//			t.Errorf("Node Summary missing for topology %s, id %s", topology, id)
//		case summary.Label == "":
//			t.Errorf("Node Summary Label missing for topology %s, id %s", topology, id)
//		case summary.Label == id && topology != processNameTopology:
//			t.Errorf("Node Summary Label same as id (that's cheating!) for topology %s, id %s", topology, id)
//		}
//	}
//}
//
//func TestNodeMetadata(t *testing.T) {
//	inputs := []struct {
//		name string
//		node report.Node
//		want []report.MetadataRow
//	}{
//		{
//			name: "container",
//			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
//				docker.ContainerID:            fixture.ClientContainerID,
//				docker.LabelPrefix + "label1": "label1value",
//				docker.ContainerStateHuman:    docker.StateRunning,
//			}).WithTopology(report.Container).WithSets(report.MakeSets().
//				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
//			),
//			want: []report.MetadataRow{
//				{ID: docker.ContainerStateHuman, Label: "State", Value: "running", Priority: 4},
//				{ID: docker.ContainerIPs, Label: "IPs", Value: "10.10.10.0/24, 10.10.10.1/24", Priority: 8},
//				{ID: docker.ContainerID, Label: "ID", Value: fixture.ClientContainerID, Priority: 11, Truncate: 12},
//			},
//		},
//		{
//			name: "unknown topology",
//			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
//				docker.ContainerID: fixture.ClientContainerID,
//			}).WithTopology("foobar"),
//			want: nil,
//		},
//	}
//	for _, input := range inputs {
//		summary, _ := detailed.MakeNodeSummary(detailed.RenderContext{Report: fixture.Report}, input.node)
//		have := summary.Metadata
//		if !reflect.DeepEqual(input.want, have) {
//			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
//		}
//	}
//}
//
//func TestNodeMetrics(t *testing.T) {
//	inputs := []struct {
//		name string
//		node report.Node
//		want []report.MetricRow
//	}{
//		{
//			name: "process",
//			node: fixture.Report.Process.Nodes[fixture.ClientProcess1NodeID],
//			want: []report.MetricRow{
//				{
//					ID:       process.CPUUsage,
//					Label:    "CPU",
//					Format:   "percent",
//					Group:    "",
//					Value:    0.01,
//					Priority: 1,
//					Metric:   &fixture.ClientProcess1CPUMetric,
//				},
//				{
//					ID:       process.MemoryUsage,
//					Label:    "Memory",
//					Format:   "filesize",
//					Group:    "",
//					Value:    0.02,
//					Priority: 2,
//					Metric:   &fixture.ClientProcess1MemoryMetric,
//				},
//			},
//		},
//		{
//			name: "container",
//			node: fixture.Report.Container.Nodes[fixture.ClientContainerNodeID],
//			want: []report.MetricRow{
//				{
//					ID:       docker.CPUTotalUsage,
//					Label:    "CPU",
//					Format:   "percent",
//					Group:    "",
//					Value:    0.03,
//					Priority: 1,
//					Metric:   &fixture.ClientContainerCPUMetric,
//				},
//				{
//					ID:       docker.MemoryUsage,
//					Label:    "Memory",
//					Format:   "filesize",
//					Group:    "",
//					Value:    0.04,
//					Priority: 2,
//					Metric:   &fixture.ClientContainerMemoryMetric,
//				},
//			},
//		},
//		{
//			name: "host",
//			node: fixture.Report.Host.Nodes[fixture.ClientHostNodeID],
//			want: []report.MetricRow{
//				{
//					ID:       host.CPUUsage,
//					Label:    "CPU",
//					Format:   "percent",
//					Group:    "",
//					Value:    0.07,
//					Priority: 1,
//					Metric:   &fixture.ClientHostCPUMetric,
//				},
//				{
//					ID:       host.MemoryUsage,
//					Label:    "Memory",
//					Format:   "filesize",
//					Group:    "",
//					Value:    0.08,
//					Priority: 2,
//					Metric:   &fixture.ClientHostMemoryMetric,
//				},
//				{
//					ID:       host.Load1,
//					Label:    "Load (1m)",
//					Group:    "load",
//					Value:    0.09,
//					Priority: 11,
//					Metric:   &fixture.ClientHostLoad1Metric,
//				},
//			},
//		},
//		{
//			name: "unknown topology",
//			node: report.MakeNode(fixture.ClientContainerNodeID).WithTopology("foobar"),
//			want: nil,
//		},
//	}
//	for _, input := range inputs {
//		summary, _ := detailed.MakeNodeSummary(detailed.RenderContext{Report: fixture.Report}, input.node)
//		have := summary.Metrics
//		if !reflect.DeepEqual(input.want, have) {
//			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
//		}
//	}
//}
//
//func TestMetricRowSummary(t *testing.T) {
//	var (
//		now    = time.Now()
//		metric = report.MakeSingletonMetric(now, 1.234)
//		row    = report.MetricRow{
//			ID:       "id",
//			Format:   "format",
//			Group:    "group",
//			Value:    1.234,
//			Priority: 1,
//			Metric:   &metric,
//		}
//		summary = row.Summary()
//	)
//	// summary should not have any samples
//	if summary.Metric.Len() != 0 {
//		t.Errorf("Expected summary to have no samples, but had %d", summary.Metric.Len())
//	}
//	// original metric should still have its samples
//	if metric.Len() != 1 {
//		t.Errorf("Expected original metric to still have it's samples, but had %d", metric.Len())
//	}
//	// summary should have all the same fields (minus the metric)
//	summary.Metric = nil
//	row.Metric = nil
//	if !reflect.DeepEqual(summary, row) {
//		t.Errorf("Expected summary to have same fields as original: %s", test.Diff(summary, row))
//	}
//}
//
//func TestNodeTables(t *testing.T) {
//	inputs := []struct {
//		name string
//		rpt  report.Report
//		node report.Node
//		want []report.Table
//	}{
//		{
//			name: "container",
//			rpt: report.Report{
//				Container: report.MakeTopology(),//.
//					//WithTableTemplates(docker.ContainerTableTemplates),
//			},
//			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
//				docker.ContainerID:            fixture.ClientContainerID,
//				docker.LabelPrefix + "label1": "label1value",
//				docker.ContainerState:         docker.StateRunning,
//			}).WithTopology(report.Container).WithSets(report.MakeSets().
//				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
//			),
//			want: []report.Table{
//				{
//					ID:    docker.EnvPrefix,
//					Type:  report.PropertyListType,
//					Label: "Environment variables",
//					Rows:  []report.Row{},
//				},
//				{
//					ID:    docker.LabelPrefix,
//					Type:  report.PropertyListType,
//					Label: "Docker labels",
//					Rows: []report.Row{
//						{
//							ID: "label_label1",
//							Entries: map[string]string{
//								"label": "label1",
//								"value": "label1value",
//							},
//						},
//					},
//				},
//				{
//					ID:    docker.ImageTableID,
//					Type:  report.PropertyListType,
//					Label: "Image",
//					Rows:  []report.Row{},
//				},
//			},
//		},
//		{
//			name: "unknown topology",
//			rpt:  report.MakeReport(),
//			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
//				docker.ContainerID: fixture.ClientContainerID,
//			}).WithTopology("foobar"),
//			want: nil,
//		},
//	}
//	for _, input := range inputs {
//		summary, _ := detailed.MakeNodeSummary(detailed.RenderContext{Report: input.rpt}, input.node)
//		have := summary.Tables
//		if !reflect.DeepEqual(input.want, have) {
//			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
//		}
//	}
//}
