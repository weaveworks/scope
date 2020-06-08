package render_test

import (
	"context"
	"testing"
	"time"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestPropagateSingleMetrics(t *testing.T) {
	now := time.Now()
	empty := report.MakeNode("empty")
	child1 := report.MakeNode(report.MakeContainerNodeID("child1")).
		WithTopology(report.Container).
		WithMetrics(report.Metrics{
			"metric1": report.MakeMetric(nil),
		})
	child2 := report.MakeNode(report.MakeContainerNodeID("child2")).
		WithTopology(report.Container).
		WithMetrics(report.Metrics{
			"metric2": report.MakeMetric(nil),
		})
	child2p := report.MakeNode(report.MakeProcessNodeID("host1", "child2")).
		WithTopology("Process").
		WithMetrics(report.Metrics{
			"metric2": report.MakeMetric(nil),
		})
	pause := report.MakeNode(report.MakeContainerNodeID("pause")).
		WithLatest(report.DoesNotMakeConnections, now, "").
		WithTopology(report.Container).
		WithMetrics(report.Metrics{
			"metric2": report.MakeMetric(nil),
		})
	a := report.MakeNode("a").WithChildID(child1.ID)
	rpt := report.MakeReport()
	rpt.Container.AddNode(empty)
	rpt.Container.AddNode(a)
	rpt.Container.AddNode(child1)
	rpt.Container.AddNode(child2)
	rpt.Container.AddNode(pause)
	rpt.Process.AddNode(child2p)
	for _, c := range []struct {
		name     string
		input    report.Node
		topology string
		output   report.Nodes
	}{
		{
			name:     "empty",
			input:    empty,
			topology: "",
			output:   report.Nodes{"empty": report.MakeNode("empty")},
		},
		{
			name:     "one child",
			input:    a,
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildID(child1.ID),
			},
		},
		{
			name:     "ignores other topologies",
			input:    report.MakeNode("a").WithChildren(report.MakeIDList(child1.ID, child2p.ID)),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildren(report.MakeIDList(child1.ID, child2p.ID)),
			},
		},
		{
			name:     "two children",
			input:    report.MakeNode("a").WithChildren(report.MakeIDList(child1.ID, child2.ID)),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithChildren(report.MakeIDList(child1.ID, child2.ID)),
			},
		},
		{
			name:     "ignores k8s pause container",
			input:    report.MakeNode("a").WithChildren(report.MakeIDList(child1.ID, pause.ID)),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildren(report.MakeIDList(child1.ID, pause.ID)),
			},
		},
	} {
		got := render.PropagateSingleMetrics(c.topology, mockRenderer{report.Nodes{c.input.ID: c.input}}).Render(context.Background(), rpt).Nodes
		if !reflect.DeepEqual(got, c.output) {
			t.Errorf("[%s] Diff: %s", c.name, test.Diff(c.output, got))
		}
	}
}
