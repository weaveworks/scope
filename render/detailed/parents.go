package detailed

import (
	"sort"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// Parent is the information needed to build a link to the parent of a Node.
type Parent struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TopologyID string `json:"topologyId"`
}

// parents renders the parents of this report.Node, which have been aggregated
// from the probe reports.
func parents(r report.Report, n render.RenderableNode) (result []Parent) {
	topologies := map[string]struct {
		report.Topology
		render func(report.Node) Parent
	}{
		report.Container:      {r.Container, containerParent},
		report.Pod:            {r.Pod, podParent},
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
		parents, _ := n.Node.Parents.Lookup(topologyID)
		for _, id := range parents {
			if topologyID == n.Node.Topology && id == n.ID {
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
	label, _ := render.GetRenderableContainerName(n)
	containerID, _ := n.Latest.Lookup(docker.ContainerID)
	return Parent{
		ID:         render.MakeContainerID(containerID),
		Label:      label,
		TopologyID: "containers",
	}
}

func podParent(n report.Node) Parent {
	podID, _ := n.Latest.Lookup(kubernetes.PodID)
	podName, _ := n.Latest.Lookup(kubernetes.PodName)
	return Parent{
		ID:         render.MakePodID(podID),
		Label:      podName,
		TopologyID: "pods",
	}
}

func serviceParent(n report.Node) Parent {
	serviceID, _ := n.Latest.Lookup(kubernetes.ServiceID)
	serviceName, _ := n.Latest.Lookup(kubernetes.ServiceName)
	return Parent{
		ID:         render.MakeServiceID(serviceID),
		Label:      serviceName,
		TopologyID: "pods-by-service",
	}
}

func containerImageParent(n report.Node) Parent {
	imageName, _ := n.Latest.Lookup(docker.ImageName)
	return Parent{
		ID:         render.MakeContainerImageID(render.ImageNameWithoutVersion(imageName)),
		Label:      imageName,
		TopologyID: "containers-by-image",
	}
}

func hostParent(n report.Node) Parent {
	hostName, _ := n.Latest.Lookup(host.HostName)
	return Parent{
		ID:         render.MakeHostID(hostName),
		Label:      hostName,
		TopologyID: "hosts",
	}
}
