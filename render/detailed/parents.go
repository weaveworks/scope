package detailed

import (
	"github.com/weaveworks/scope/report"
)

// Parent is the information needed to build a link to the parent of a Node.
type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

// parent topologies, in the order we want to show them
var parentTopologies = []string{
	report.Container,
	report.ContainerImage,
	report.Pod,
	report.Deployment,
	report.DaemonSet,
	report.StatefulSet,
	report.CronJob,
	report.Service,
	report.ECSTask,
	report.ECSService,
	report.SwarmService,
	report.Host,
}

// Parents renders the parents of this report.Node, which have been aggregated
// from the probe reports.
func Parents(r report.Report, n report.Node) []Parent {
	if n.Parents.Size() == 0 {
		return nil
	}
	result := make([]Parent, 0, n.Parents.Size())
	for _, topologyID := range parentTopologies {
		topology, ok := r.Topology(topologyID)
		if !ok {
			continue
		}
		apiTopologyID, ok := primaryAPITopology[topologyID]
		if !ok {
			continue
		}
		parents, _ := n.Parents.Lookup(topologyID)
		for _, id := range parents {
			if topologyID == n.Topology && id == n.ID {
				continue
			}
			parentNode, ok := topology.Nodes[id]
			if !ok {
				parentNode = report.MakeNode(id).WithTopology(topologyID)
			}
			if summary, ok := MakeBasicNodeSummary(r, parentNode); ok {
				result = append(result, Parent{
					ID:         summary.ID,
					Label:      summary.Label,
					TopologyID: apiTopologyID,
				})
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
