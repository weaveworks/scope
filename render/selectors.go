package render

import (
	"github.com/weaveworks/scope/report"
)

// TopologySelector selects a single topology from a report.
// NB it is also a Renderer!
type TopologySelector string

// Render implements Renderer
func (t TopologySelector) Render(r report.Report, _ Decorator) report.Nodes {
	topology, _ := r.Topology(string(t))
	return topology.Nodes
}

// Stats implements Renderer
func (t TopologySelector) Stats(r report.Report, _ Decorator) Stats {
	return Stats{}
}

// The topology selectors implement a Renderer which fetch the nodes from the
// various report topologies.
var (
	SelectEndpoint       = TopologySelector(report.Endpoint)
	SelectProcess        = TopologySelector(report.Process)
	SelectContainer      = TopologySelector(report.Container)
	SelectContainerImage = TopologySelector(report.ContainerImage)
	SelectHost           = TopologySelector(report.Host)
	SelectPod            = TopologySelector(report.Pod)
	SelectService        = TopologySelector(report.Service)
)
