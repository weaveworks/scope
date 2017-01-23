package detailed

import (
	"sort"

	"github.com/weaveworks/scope/probe/awsecs"
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

func node(t report.Topology) func(string) (report.Node, bool) {
	return func(id string) (report.Node, bool) {
		n, ok := t.Nodes[id]
		return n, ok
	}
}

func fake(id string) (report.Node, bool) {
	return report.MakeNode(id), true
}

// Parents renders the parents of this report.Node, which have been aggregated
// from the probe reports.
func Parents(r report.Report, n report.Node) (result []Parent) {
	topologies := map[string]struct {
		node   func(id string) (report.Node, bool)
		render func(report.Node) Parent
	}{
		report.Container:      {node(r.Container), containerParent},
		report.Pod:            {node(r.Pod), podParent},
		report.ReplicaSet:     {node(r.ReplicaSet), replicaSetParent},
		report.Deployment:     {node(r.Deployment), deploymentParent},
		report.Service:        {node(r.Service), serviceParent},
		report.ECSTask:        {node(r.ECSTask), ecsTaskParent},
		report.ECSService:     {node(r.ECSService), ecsServiceParent},
		report.ContainerImage: {fake, containerImageParent},
		report.Host:           {node(r.Host), hostParent},
	}
	topologyIDs := []string{}
	for topologyID := range topologies {
		topologyIDs = append(topologyIDs, topologyID)
	}
	sort.Strings(topologyIDs)
	for _, topologyID := range topologyIDs {
		t := topologies[topologyID]
		parents, _ := n.Parents.Lookup(topologyID)
		for _, id := range parents {
			if topologyID == n.Topology && id == n.ID {
				continue
			}

			parent, ok := t.node(id)
			if !ok {
				continue
			}

			result = append(result, t.render(parent))
		}
	}
	return result
}

func containerParent(n report.Node) Parent {
	label := getRenderableContainerName(n)
	return Parent{
		ID:         n.ID,
		Label:      label,
		TopologyID: "containers",
	}
}

var (
	podParent        = kubernetesParent("pods")
	replicaSetParent = kubernetesParent("replica-sets")
	deploymentParent = kubernetesParent("deployments")
	serviceParent    = kubernetesParent("services")
)

func kubernetesParent(topology string) func(report.Node) Parent {
	return func(n report.Node) Parent {
		name, _ := n.Latest.Lookup(kubernetes.Name)
		return Parent{
			ID:         n.ID,
			Label:      name,
			TopologyID: topology,
		}
	}
}

func ecsTaskParent(n report.Node) Parent {
	family, _ := n.Latest.Lookup(awsecs.TaskFamily)
	return Parent{
		ID:         n.ID,
		Label:      family,
		TopologyID: "ecs-tasks",
	}
}

func ecsServiceParent(n report.Node) Parent {
	name, _ := report.ParseECSServiceNodeID(n.ID)
	return Parent{
		ID:         n.ID,
		Label:      name,
		TopologyID: "ecs-services",
	}
}

func containerImageParent(n report.Node) Parent {
	name, _ := report.ParseContainerImageNodeID(n.ID)
	return Parent{
		ID:         n.ID,
		Label:      name,
		TopologyID: "containers-by-image",
	}
}

func hostParent(n report.Node) Parent {
	hostName, _ := n.Latest.Lookup(host.HostName)
	return Parent{
		ID:         n.ID,
		Label:      hostName,
		TopologyID: "hosts",
	}
}
