package render

import (
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UnmanagedID    = "unmanaged"
	UnmanagedMajor = "Unmanaged"
)

// UnmanagedIDPrefix is the prefix of unmanaged pseudo nodes
var UnmanagedIDPrefix = MakePseudoNodeID(UnmanagedID)

func renderKubernetesTopologies(rpt report.Report) bool {
	// Render if any k8s topology has any nodes
	topologies := []*report.Topology{
		&rpt.Pod,
		&rpt.Service,
		&rpt.Deployment,
		&rpt.DaemonSet,
		&rpt.StatefulSet,
		&rpt.CronJob,
	}
	for _, t := range topologies {
		if len(t.Nodes) > 0 {
			return true
		}
	}
	return false
}

func isPauseContainer(n report.Node) bool {
	image, ok := n.Latest.Lookup(docker.ImageName)
	return ok && kubernetes.IsPauseImageName(image)
}

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = Memoise(ConditionalRenderer(renderKubernetesTopologies,
	MakeFilter(
		func(n report.Node) bool {
			state, ok := n.Latest.Lookup(kubernetes.State)
			return (!ok || state != kubernetes.StateDeleted)
		},
		MakeReduce(
			MakeMap(
				PropagateSingleMetrics(report.Container),
				MakeMap(
					Map2Parent([]string{report.Pod}, UnmanagedID),
					MakeFilter(
						ComposeFilterFuncs(
							IsRunning,
							Complement(isPauseContainer),
						),
						ContainerWithImageNameRenderer,
					),
				),
			),
			// ConnectionJoin invokes the renderer twice, hence it
			// helps to memoise it.
			ConnectionJoin(MapPod2IP, Memoise(selectPodsWithDeployments{})),
		),
	),
))

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
//
// not memoised
var PodServiceRenderer = ConditionalRenderer(renderKubernetesTopologies,
	renderParents(
		report.Pod, []string{report.Service}, "",
		PodRenderer,
	),
)

// KubeControllerRenderer is a Renderer which combines all the 'controller' topologies.
// Pods with no controller are mapped to 'Unmanaged'
// We can't simply combine the rendered graphs of the high level objects as they would never
// have connections to each other.
//
// not memoised
var KubeControllerRenderer = ConditionalRenderer(renderKubernetesTopologies,
	renderParents(
		report.Pod, []string{report.Deployment, report.DaemonSet, report.StatefulSet, report.CronJob}, UnmanagedID,
		PodRenderer,
	),
)

// renderParents produces a 'standard' renderer for mapping from some child topology to some parent topologies,
// by taking a child renderer, mapping to parents, propagating single metrics, and joining with full parent topology.
// Other options are as per Map2Parent.
func renderParents(childTopology string, parentTopologies []string, noParentsPseudoID string, childRenderer Renderer) Renderer {
	selectors := make([]Renderer, len(parentTopologies))
	for i, topology := range parentTopologies {
		selectors[i] = TopologySelector(topology)
	}
	return MakeReduce(append(
		selectors,
		MakeMap(
			PropagateSingleMetrics(childTopology),
			MakeMap(
				Map2Parent(parentTopologies, noParentsPseudoID),
				childRenderer,
			),
		),
	)...)
}

// Renderer to return modified Pod nodes to elide replica sets and point directly
// to deployments where applicable.
type selectPodsWithDeployments struct{}

func (s selectPodsWithDeployments) Render(rpt report.Report, dct Decorator) Nodes {
	result := report.Nodes{}
	// For each pod, we check for any replica sets, and merge any deployments they point to
	// into a replacement Parents value.
	for podID, pod := range rpt.Pod.Nodes {
		if replicaSetIDs, ok := pod.Parents.Lookup(report.ReplicaSet); ok {
			newParents := pod.Parents.Delete(report.ReplicaSet)
			for _, replicaSetID := range replicaSetIDs {
				if replicaSet, ok := rpt.ReplicaSet.Nodes[replicaSetID]; ok {
					if deploymentIDs, ok := replicaSet.Parents.Lookup(report.Deployment); ok {
						newParents = newParents.Add(report.Deployment, deploymentIDs)
					}
				}
			}
			pod = pod.WithParents(newParents)
		}
		result[podID] = pod
	}
	return Nodes{Nodes: result}
}

func (s selectPodsWithDeployments) Stats(rpt report.Report, _ Decorator) Stats {
	return Stats{}
}

// MapPod2IP maps pod nodes to their IP address.  This allows pods to
// be joined directly with the endpoint topology.
func MapPod2IP(m report.Node) []string {
	// if this pod belongs to the host's networking namespace
	// we cannot use its IP to attribute connections
	// (they could come from any other process on the host or DNAT-ed IPs)
	if _, ok := m.Latest.Lookup(kubernetes.IsInHostNetwork); ok {
		return nil
	}

	ip, ok := m.Latest.Lookup(kubernetes.IP)
	if !ok {
		return nil
	}
	return []string{report.MakeScopedEndpointNodeID("", ip, "")}
}

// Map2Parent returns a MapFunc which maps Nodes to some parent grouping.
func Map2Parent(
	// The topology IDs to look for parents in
	topologies []string,
	// Either the ID prefix of the pseudo node to use for nodes without
	// any parents in the group, eg. UnmanagedID, or "" to drop nodes without any parents.
	noParentsPseudoID string,
) MapFunc {
	return func(n report.Node, _ report.Networks) report.Nodes {
		// Uncontained becomes Unmanaged/whatever if noParentsPseudoID is set
		if noParentsPseudoID != "" && strings.HasPrefix(n.ID, UncontainedIDPrefix) {
			id := MakePseudoNodeID(noParentsPseudoID, report.ExtractHostID(n))
			node := NewDerivedPseudoNode(id, n)
			return report.Nodes{id: node}
		}

		// Propagate all pseudo nodes
		if n.Topology == Pseudo {
			return report.Nodes{n.ID: n}
		}

		// For each topology, map to any parents we can find
		result := report.Nodes{}
		for _, topology := range topologies {
			if groupIDs, ok := n.Parents.Lookup(topology); ok {
				for _, id := range groupIDs {
					node := NewDerivedNode(id, n).WithTopology(topology)
					node.Counters = node.Counters.Add(n.Topology, 1)
					result[id] = node
				}
			}
		}

		if len(result) == 0 && noParentsPseudoID != "" {
			// Map to pseudo node
			id := MakePseudoNodeID(noParentsPseudoID, report.ExtractHostID(n))
			node := NewDerivedPseudoNode(id, n)
			result[id] = node
		}

		return result
	}
}
