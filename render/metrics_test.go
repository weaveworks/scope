package render_test

import (
	"testing"
	"time"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
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
							"metric1": report.MakeMetric(nil),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(nil),
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
							"metric1": report.MakeMetric(nil),
						}),
					report.MakeNode("child2").
						WithTopology("otherTopology").
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(nil),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(nil),
							}),
						report.MakeNode("child2").
							WithTopology("otherTopology").
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(nil),
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
							"metric1": report.MakeMetric(nil),
						}),
					report.MakeNode("child2").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(nil),
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
								"metric1": report.MakeMetric(nil),
							}),
						report.MakeNode("child2").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(nil),
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
							"metric1": report.MakeMetric(nil),
						}),
					report.MakeNode("child2").
						WithLatest(report.DoesNotMakeConnections, now, "").
						WithTopology(report.Container).
						WithMetrics(report.Metrics{
							"metric2": report.MakeMetric(nil),
						}),
				),
			),
			topology: report.Container,
			output: report.Nodes{
				"a": report.MakeNode("a").WithMetrics(report.Metrics{
					"metric1": report.MakeMetric(nil),
				}).WithChildren(
					report.MakeNodeSet(
						report.MakeNode("child1").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric1": report.MakeMetric(nil),
							}),
						report.MakeNode("child2").
							WithLatest(report.DoesNotMakeConnections, now, "").
							WithTopology(report.Container).
							WithMetrics(report.Metrics{
								"metric2": report.MakeMetric(nil),
							}),
					),
				),
			},
		},
	} {
		got := render.PropagateSingleMetrics(c.topology)(c.input)
		if !reflect.DeepEqual(got, c.output) {
			t.Errorf("[%s] Diff: %s", c.name, test.Diff(c.output, got))
		}
	}
}
