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
				docker.ContainerStateHuman:    docker.StateRunning,
			}).WithTopology(report.Container).WithSets(report.EmptySets.
				Add(docker.ContainerIPs, report.MakeStringSet("10.10.10.0/24", "10.10.10.1/24")),
			),
			want: []detailed.MetadataRow{
				{ID: docker.ContainerID, Value: fixture.ClientContainerID, Prime: true},
				{ID: docker.ContainerStateHuman, Value: "running", Prime: true},
				{ID: docker.ContainerIPs, Value: "10.10.10.0/24, 10.10.10.1/24"},
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

func TestMetadataRowCopy(t *testing.T) {
	var (
		row = detailed.MetadataRow{
			ID:       "id",
			Value:    "value",
			Prime:    true,
			Datatype: "datatype",
		}
		cp = row.Copy()
	)

	// copy should be identical
	if !reflect.DeepEqual(row, cp) {
		t.Error(test.Diff(row, cp))
	}

	// changing the copy should not change the original
	cp.ID = ""
	cp.Value = ""
	cp.Prime = false
	cp.Datatype = ""
	if row.ID != "id" || row.Value != "value" || row.Prime != true || row.Datatype != "datatype" {
		t.Errorf("Expected changing the copy not to modify the original")
	}
}
