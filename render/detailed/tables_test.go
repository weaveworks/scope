package detailed_test

import (
	"reflect"
	"testing"

	"$GITHUB_URI/probe/docker"
	"$GITHUB_URI/render/detailed"
	"$GITHUB_URI/report"
	"$GITHUB_URI/test"
	"$GITHUB_URI/test/fixture"
)

func TestNodeTables(t *testing.T) {
	inputs := []struct {
		name string
		rpt  report.Report
		node report.Node
		want []report.Table
	}{
		{
			name: "container",
			rpt: report.Report{
				Container: report.MakeTopology().
					WithTableTemplates(docker.ContainerTableTemplates),
			},
			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
				docker.ContainerID:            fixture.ClientContainerID,
				docker.LabelPrefix + "label1": "label1value",
				docker.ContainerState:         docker.StateRunning,
			}).WithTopology(report.Container).WithSets(report.EmptySets.
				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
			),
			want: []report.Table{
				{
					ID:    docker.EnvPrefix,
					Label: "Environment Variables",
					Rows:  []report.MetadataRow{},
				},
				{
					ID:    docker.LabelPrefix,
					Label: "Docker Labels",
					Rows: []report.MetadataRow{
						{
							ID:    "label_label1",
							Label: "label1",
							Value: "label1value",
						},
					},
				},
			},
		},
		{
			name: "unknown topology",
			rpt:  report.MakeReport(),
			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
				docker.ContainerID: fixture.ClientContainerID,
			}).WithTopology("foobar"),
			want: nil,
		},
	}
	for _, input := range inputs {
		have := detailed.NodeTables(input.rpt, input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}
