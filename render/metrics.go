package render

import (
	"github.com/weaveworks/scope/report"
)

// PropagateSingleMetrics creates a renderer which propagates metrics
// from a node's child to the node. The child is selected based on the
// specified topology. No metrics are propagated when there is more
// than one such child.
func PropagateSingleMetrics(topology string, r Renderer) Renderer {
	return propagateSingleMetrics{topology: topology, r: r}
}

type propagateSingleMetrics struct {
	topology string
	r        Renderer
}

func (p propagateSingleMetrics) Render(rpt report.Report) Nodes {
	nodes := p.r.Render(rpt)
	outputs := make(report.Nodes, len(nodes.Nodes))
	for id, n := range nodes.Nodes {
		var found []report.Node
		n.Children.ForEach(func(child report.Node) {
			if child.Topology == p.topology {
				if _, ok := child.Latest.Lookup(report.DoesNotMakeConnections); !ok {
					found = append(found, child)
				}
			}
		})
		if len(found) == 1 {
			n = n.WithMetrics(found[0].Metrics)
		}
		outputs[id] = n
	}
	return Nodes{Nodes: outputs, Filtered: nodes.Filtered}
}
