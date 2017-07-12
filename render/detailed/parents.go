package detailed

import (
	"sort"

	"github.com/weaveworks/scope/probe/awsecs"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// Parent is the information needed to build a link to the parent of a Node.
type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

var (
	kubernetesParentLabel = latestLookup(kubernetes.Name)

	getLabelForTopology = map[string]func(report.Node) string{
		report.Container:      getRenderableContainerName,
		report.Pod:            kubernetesParentLabel,
		report.Deployment:     kubernetesParentLabel,
		report.DaemonSet:      kubernetesParentLabel,
		report.StatefulSet:    kubernetesParentLabel,
		report.CronJob:        kubernetesParentLabel,
		report.Service:        kubernetesParentLabel,
		report.ECSTask:        latestLookup(awsecs.TaskFamily),
		report.ECSService:     ecsServiceParentLabel,
		report.SwarmService:   latestLookup(docker.ServiceName),
		report.ContainerImage: containerImageParentLabel,
		report.Host:           latestLookup(host.HostName),
	}
)

// Parents renders the parents of this report.Node, which have been aggregated
// from the probe reports.
func Parents(r report.Report, n report.Node) (result []Parent) {
	topologyIDs := []string{}
	for topologyID := range getLabelForTopology {
		topologyIDs = append(topologyIDs, topologyID)
	}
	sort.Strings(topologyIDs)
	for _, topologyID := range topologyIDs {
		getLabel := getLabelForTopology[topologyID]
		topology, ok := r.Topology(topologyID)
		if !ok {
			continue
		}
		parents, _ := n.Parents.Lookup(topologyID)
		for _, id := range parents {
			if topologyID == n.Topology && id == n.ID {
				continue
			}

			var parentNode report.Node
			// Special case: container image parents should be empty nodes for some reason
			if topologyID == report.ContainerImage {
				parentNode = report.MakeNode(id)
			} else {
				if parent, ok := topology.Nodes[id]; ok {
					parentNode = parent
				} else {
					continue
				}
			}

			apiTopologyID, ok := primaryAPITopology[topologyID]
			if !ok {
				continue
			}

			result = append(result, Parent{
				ID:         id,
				Label:      getLabel(parentNode),
				TopologyID: apiTopologyID,
			})
		}
	}
	return result
}

func latestLookup(key string) func(report.Node) string {
	return func(n report.Node) string {
		value, _ := n.Latest.Lookup(key)
		return value
	}
}

func ecsServiceParentLabel(n report.Node) string {
	_, name, _ := report.ParseECSServiceNodeID(n.ID)
	return name
}

func containerImageParentLabel(n report.Node) string {
	name, _ := report.ParseContainerImageNodeID(n.ID)
	return name
}
