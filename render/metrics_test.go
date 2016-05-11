package render_test

import (
	"testing"
	"time"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestPropagateSingleMetrics(t *testing.T) {
	now := time.Now()
	for _, c := range []struct {
		name     string
		input    report.Node
		topology string
		output   report.Nodes
	}{
		{
			name:     "empty",
			input:    report.MakeNode("empty"),
			topology: "",
			output:   report.Nodes{"empty": report.MakeNode("empty")},
		},
		{
			name: "one child",
			input: report.MakeNode("a").WithChildren(
				report.MakeNodeSet(
					report.MakeNode("child1").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric1": report.MakeMetric(),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(),
							}),
					),
				),
			},
		},
		{
			name: "ignores other topologies",
			input: report.MakeNode("a").WithChildren(
				report.MakeNodeSet(
					report.MakeNode("child1").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric1": report.MakeMetric(),
						}),
					report.MakeNode("child2").
						WithTopology("otherTopology").
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(),
							}),
						report.MakeNode("child2").
							WithTopology("otherTopology").
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(),
							}),
					),
				),
			},
		},
		{
			name: "two children",
			input: report.MakeNode("a").WithChildren(
				report.MakeNodeSet(
					report.MakeNode("child1").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric1": report.MakeMetric(),
						}),
					report.MakeNode("child2").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(),
							}),
						report.MakeNode("child2").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(),
							}),
					),
				),
			},
		},
		{
			name: "ignores k8s pause container",
			input: report.MakeNode("a").WithChildren(
				report.MakeNodeSet(
					report.MakeNode("child1").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric1": report.MakeMetric(),
						}),
					report.MakeNode("child2").
						WithLatest(report.DoesNotMakeConnections, now, "").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(),
							}),
						report.MakeNode("child2").
							WithLatest(report.DoesNotMakeConnections, now, "").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(),
							}),
					),
				),
			},
		},
	} {
		got := render.PropagateSingleMetrics(c.topology)(c.input, report.Networks{})
		if !reflect.DeepEqual(got, c.output) {
			t.Errorf("[%s] Diff: %s", c.name, test.Diff(c.output, got))
		}
	}
}
