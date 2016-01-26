package detailed_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestNodeMetadata(t *testing.T) {
	inputs := []struct {
		name string
		node report.Node
		want []detailed.MetadataRow
	}{
		{
			name: "container",
			node: report.MakeNodeWith(map[string]string{
				docker.ContainerID:            fixture.ClientContainerID,
				docker.LabelPrefix + "label1": "label1value",
				docker.ContainerState:         docker.StateRunning,
			}).WithTopology(report.Container).WithSets(report.EmptySets.
				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
			),
			want: []detailed.MetadataRow{
				{ID: docker.ContainerID, Label: "ID", Value: fixture.ClientContainerID},
				{ID: docker.ContainerState, Label: "State", Value: "running"},
				{ID: docker.ContainerIPs, Label: "IPs", Value: "10.10.10.0/24, 10.10.10.1/24"},
				{
					ID:    "label_label1",
					Label: "Label \"label1\"",
					Value: "label1value",
				},
			},
		},
		{
			name: "unknown topology",
			node: report.MakeNodeWith(map[string]string{
				docker.ContainerID: fixture.ClientContainerID,
			}).WithTopology("foobar").WithID(fixture.ClientContainerNodeID),
			want: nil,
		},
	}
	for _, input := range inputs {
		have := detailed.NodeMetadata(input.node)
		if !reflect.DeepEqual(input.want, have) {
			t.Errorf("%s: %s", input.name, test.Diff(input.want, have))
		}
	}
}
