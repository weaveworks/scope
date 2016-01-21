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
			}).WithTopology(report.Container).WithSets(report.Sets{
				docker.ContainerIPs: report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24"),
			}).WithLatest(docker.ContainerState, fixture.Now, docker.StateRunning),
			want: []detailed.MetadataRow{
				{ID: docker.ContainerID, Value: fixture.ClientContainerID},
				{ID: docker.ContainerState, Value: "running"},
				{ID: docker.ContainerIPs, Value: "10.10.10.0/24, 10.10.10.1/24"},
				{
					ID:    "label_label1",
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
