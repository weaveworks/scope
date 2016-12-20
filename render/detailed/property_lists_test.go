package detailed_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

func TestNodePropertyLists(t *testing.T) {
	inputs := []struct {
		name string
		rpt  report.Report
		node report.Node
		want []report.PropertyList
	}{
		{
			name: "container",
			rpt: report.Report{
				Container: report.MakeTopology().
					WithPropertyListTemplates(docker.ContainerPropertyListTemplates),
			},
			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
				docker.ContainerID:            fixture.ClientContainerID,
				docker.LabelPrefix + "label1": "label1value",
				docker.ContainerState:         docker.StateRunning,
			}).WithTopology(report.Container).WithSets(report.EmptySets.
				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
			),
			want: []report.PropertyList{
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
				{
					ID:    docker.ImageTableID,
					Label: "Image",
					Rows:  []report.MetadataRow{},
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
		have := detailed.NodePropertyLists(input.rpt, input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}
