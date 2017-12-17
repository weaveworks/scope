package render

import (
	"github.com/weaveworks/scope/report"
)

// PropagateSingleMetrics puts metrics from one of the children onto the parent
// iff there is only one child of that type.
func PropagateSingleMetrics(topology string) MapFunc {
	return func(n report.Node) report.Nodes {
		var found []report.Node
		n.Children.ForEach(func(child report.Node) {
			if child.Topology == topology {
				if _, ok := child.Latest.Lookup(report.DoesNotMakeConnections); !ok {
					found = append(found, child)
				}
			}
		})
		if len(found) == 1 {
			n = n.WithMetrics(found[0].Metrics)
		}
		return report.Nodes{n.ID: n}
	}
}
