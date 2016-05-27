package detailed

import (
	"sort"

	"$GITHUB_URI/probe/docker"
	"$GITHUB_URI/probe/host"
	"$GITHUB_URI/probe/kubernetes"
	"$GITHUB_URI/report"
)

// Parent is the information needed to build a link to the parent of a Node.
type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

// Parents renders the parents of this report.Node, which have been aggregated
// from the probe reports.
func Parents(r report.Report, n report.Node) (result []Parent) {
	topologies := map[string]struct {
		report.Topology
		render func(report.Node) Parent
	}{
		report.Container:      {r.Container, containerParent},
		report.Pod:            {r.Pod, podParent},
		report.ReplicaSet:     {r.ReplicaSet, replicaSetParent},
		report.Deployment:     {r.Deployment, deploymentParent},
		report.Service:        {r.Service, serviceParent},
		report.ContainerImage: {r.ContainerImage, containerImageParent},
		report.Host:           {r.Host, hostParent},
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

			parent, ok := t.Nodes[id]
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

func containerImageParent(n report.Node) Parent {
	imageName, _ := n.Latest.Lookup(docker.ImageName)
	return Parent{
		ID:         n.ID,
		Label:      docker.ImageNameWithoutVersion(imageName),
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
