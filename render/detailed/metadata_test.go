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

func TestNodeMetadata(t *testing.T) {
	inputs := []struct {
		name string
		node report.Node
		want []report.MetadataRow
	}{
		{
			name: "container",
			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
				docker.ContainerID:            fixture.ClientContainerID,
				docker.LabelPrefix + "label1": "label1value",
				docker.ContainerStateHuman:    docker.StateRunning,
			}).WithTopology(report.Container).WithSets(report.MakeSets().
				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
			),
			want: []report.MetadataRow{
				{ID: docker.ContainerStateHuman, Label: "State", Value: "running", Priority: 3},
				{ID: docker.ContainerIPs, Label: "IPs", Value: "10.10.10.0/24, 10.10.10.1/24", Priority: 7},
				{ID: docker.ContainerID, Label: "ID", Value: fixture.ClientContainerID, Priority: 10, Truncate: 12},
			},
		},
		{
			name: "unknown topology",
			node: report.MakeNodeWith(fixture.ClientContainerNodeID, map[string]string{
				docker.ContainerID: fixture.ClientContainerID,
			}).WithTopology("foobar"),
			want: nil,
		},
	}
	for _, input := range inputs {
		have := detailed.NodeMetadata(fixture.Report, input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}
