package detailed_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestParents(t *testing.T) {
	ctx := context.Background()
	for _, c := range []struct {
		name string
		node report.Node
		want []detailed.Parent
	}{
		{
			name: "Node accidentally tagged with itself",
			node: render.HostRenderer.Render(ctx, fixture.Report).Nodes[fixture.ClientHostNodeID].WithParents(
				report.MakeSets().Add(report.Host, report.MakeStringSet(fixture.ClientHostNodeID)),
			),
			want: nil,
		},
		{
			node: render.HostRenderer.Render(ctx, fixture.Report).Nodes[fixture.ClientHostNodeID],
			want: nil,
		},
		{
			name: "Container image",
			node: render.ContainerImageRenderer.Render(ctx, fixture.Report).Nodes[expected.ClientContainerImageNodeID],
			want: []detailed.Parent{
				{ID: fixture.ClientHostNodeID, Label: "client", TopologyID: "hosts"},
			},
		},
		{
			name: "Container",
			node: render.ContainerWithImageNameRenderer.Render(ctx, fixture.Report).Nodes[fixture.ClientContainerNodeID],
			want: []detailed.Parent{
				{ID: expected.ClientContainerImageNodeID, Label: fixture.ClientContainerImageName, TopologyID: "containers-by-image"},
				{ID: fixture.ClientPodNodeID, Label: "pong-a", TopologyID: "pods"},
				{ID: fixture.ClientHostNodeID, Label: "client", TopologyID: "hosts"},
			},
		},
		{
			node: render.ProcessRenderer.Render(ctx, fixture.Report).Nodes[fixture.ClientProcess1NodeID],
			want: []detailed.Parent{
				{ID: fixture.ClientContainerNodeID, Label: fixture.ClientContainerName, TopologyID: "containers"},
				{ID: fixture.ClientHostNodeID, Label: "client", TopologyID: "hosts"},
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
