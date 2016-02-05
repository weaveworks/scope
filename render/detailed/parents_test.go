package detailed_test

import (
	"fmt"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestParents(t *testing.T) {
	for _, c := range []struct {
		name string
		node render.RenderableNode
		want []detailed.Parent
	}{
		{
			name: "Node accidentally tagged with itself",
			node: render.HostRenderer.Render(fixture.Report)[render.MakeHostID(fixture.ClientHostID)].WithParents(
				report.EmptySets.Add(report.Host, report.MakeStringSet(fixture.ClientHostNodeID)),
			),
			want: nil,
		},
		{
			node: render.HostRenderer.Render(fixture.Report)[render.MakeHostID(fixture.ClientHostID)],
			want: nil,
		},
		{
			node: render.ContainerRenderer.Render(fixture.Report)[render.MakeContainerID(fixture.ClientContainerID)],
			want: []detailed.Parent{
				{ID: render.MakeContainerImageID(fixture.ClientContainerImageName), Label: fixture.ClientContainerImageName, TopologyID: "containers-by-image"},
				{ID: render.MakeHostID(fixture.ClientHostID), Label: fixture.ClientHostName, TopologyID: "hosts"},
			},
		},
		{
			node: render.ProcessRenderer.Render(fixture.Report)[render.MakeProcessID(fixture.ClientHostID, fixture.Client1PID)],
			want: []detailed.Parent{
				{ID: render.MakeContainerID(fixture.ClientContainerID), Label: fixture.ClientContainerName, TopologyID: "containers"},
				{ID: render.MakeContainerImageID(fixture.ClientContainerImageName), Label: fixture.ClientContainerImageName, TopologyID: "containers-by-image"},
				{ID: render.MakeHostID(fixture.ClientHostID), Label: fixture.ClientHostName, TopologyID: "hosts"},
			},
		},
	} {
		name := c.name
		if name == "" {
			name = fmt.Sprintf("Node %q", c.node.ID)
		}
		if have := detailed.Parents(fixture.Report, c.node); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: %s", name, test.Diff(c.want, have))
		}
	}
}
