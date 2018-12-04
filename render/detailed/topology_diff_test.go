package detailed_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render/detailed"
	"github.com/weaveworks/scope/report"
)

// ByID is a sort interface for a NodeSummary slice.
type ByID []detailed.NodeSummary

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].ID < r[j].ID }

func TestTopoDiff(t *testing.T) {
	nodea := detailed.NodeSummary{
		BasicNodeSummary: detailed.BasicNodeSummary{
			ID:         "nodea",
			Label:      "Node A",
			LabelMinor: "'ts an a",
			Pseudo:     false,
		},
		Adjacency: report.MakeIDList("nodeb"),
	}
	nodeap := nodea
	nodeap.Adjacency = report.MakeIDList("nodeb", "nodeq") // not the same anymore
	nodeb := detailed.NodeSummary{
		BasicNodeSummary: detailed.BasicNodeSummary{
			ID:    "nodeb",
			Label: "Node B",
		},
	}

	// Helper to make RenderableNode maps.
	nodes := func(ns ...detailed.NodeSummary) detailed.NodeSummaries {
		r := detailed.NodeSummaries{}
		for _, n := range ns {
			r[n.ID] = n
		}
		return r
	}

	for _, c := range []struct {
		label      string
		have, want detailed.Diff
	}{
		{
			label: "basecase: empty -> something",
			have:  detailed.TopoDiff(nodes(), nodes(nodea, nodeb)),
			want: detailed.Diff{
				Add: []detailed.NodeSummary{nodea, nodeb},
			},
		},
		{
			label: "basecase: something -> empty",
			have:  detailed.TopoDiff(nodes(nodea, nodeb), nodes()),
			want: detailed.Diff{
				Remove: []string{"nodea", "nodeb"},
			},
		},
		{
			label: "add and remove",
			have:  detailed.TopoDiff(nodes(nodea), nodes(nodeb)),
			want: detailed.Diff{
				Add:    []detailed.NodeSummary{nodeb},
				Remove: []string{"nodea"},
			},
		},
		{
			label: "no change",
			have:  detailed.TopoDiff(nodes(nodea), nodes(nodea)),
			want:  detailed.Diff{},
		},
		{
			label: "change a single node",
			have:  detailed.TopoDiff(nodes(nodea), nodes(nodeap)),
			want: detailed.Diff{
				Update: []detailed.NodeSummary{nodeap},
			},
		},
		{
			label: "no previous nodes",
			have:  detailed.TopoDiff(nil, nodes(nodea)),
			want: detailed.Diff{
				Add:   []detailed.NodeSummary{nodea},
				Reset: true,
			},
		},
		{
			label: "empty previous nodes",
			have:  detailed.TopoDiff(nodes(), nodes(nodea)),
			want: detailed.Diff{
				Add:   []detailed.NodeSummary{nodea},
				Reset: false,
			},
		},
	} {
		sort.Strings(c.have.Remove)
		sort.Sort(ByID(c.have.Add))
		sort.Sort(ByID(c.have.Update))
		if !reflect.DeepEqual(c.want, c.have) {
			t.Errorf("%s - %s", c.label, test.Diff(c.want, c.have))
		}
	}
}
