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
var UnmanagedIDPrefix = MakePseudoNodeID(UnmanagedID, "")

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
			return !ok || !(state == kubernetes.StateDeleted || state == kubernetes.StateFailed)
		},
		MakeReduce(
			PropagateSingleMetrics(report.Container,
				Map2Parent{topologies: []string{report.Pod}, noParentsPseudoID: UnmanagedID,
					chainRenderer: MakeFilter(
						ComposeFilterFuncs(
							IsRunning,
							Complement(isPauseContainer),
						),
						ContainerWithImageNameRenderer,
					)},
			),
			ConnectionJoin(MapPod2IP, report.Pod),
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
		PropagateSingleMetrics(childTopology,
			Map2Parent{topologies: parentTopologies, noParentsPseudoID: noParentsPseudoID,
				chainRenderer: childRenderer},
		),
	)...)
}

// MapPod2IP maps pod nodes to their IP address.  This allows pods to
// be joined directly with the endpoint topology.
func MapPod2IP(_ report.Report, m report.Node) []string {
	// if this pod belongs to the host's networking namespace
	// we cannot use its IP to attribute connections
	// (they could come from any other process on the host or DNAT-ed IPs)
	if _, ok := m.Latest.Lookup(kubernetes.IsInHostNetwork); ok {
		return nil
	}

	ip, ok := m.Latest.Lookup(kubernetes.IP)
	if !ok || ip == "" {
		return nil
	}
	return []string{report.MakeScopedEndpointNodeID("", ip, "")}
}

// Map2Parent is a Renderer which maps Nodes to some parent grouping.
type Map2Parent struct {
	// Renderer to chain from
	chainRenderer Renderer
	// The topology IDs to look for parents in
	topologies []string
	// Either the ID prefix of the pseudo node to use for nodes without
	// any parents in the group, eg. UnmanagedID, or "" to drop nodes without any parents.
	noParentsPseudoID string
}

// Render implements Renderer
func (m Map2Parent) Render(rpt report.Report) Nodes {
	input := m.chainRenderer.Render(rpt)
	ret := newJoinResults(nil)

	for _, n := range input.Nodes {
		// Uncontained becomes Unmanaged/whatever if noParentsPseudoID is set
		if m.noParentsPseudoID != "" && strings.HasPrefix(n.ID, UncontainedIDPrefix) {
			id := MakePseudoNodeID(m.noParentsPseudoID, n.ID[len(UncontainedIDPrefix):])
			ret.addChildAndChildren(n, id, Pseudo)
			continue
		}

		// Propagate all pseudo nodes
		if n.Topology == Pseudo {
			ret.passThrough(n)
			continue
		}

		added := false
		// For each topology, map to any parents we can find
		for _, topology := range m.topologies {
			if groupIDs, ok := n.Parents.Lookup(topology); ok {
				for _, id := range groupIDs {
					ret.addChildAndChildren(n, id, topology)
					added = true
				}
			}
		}

		if !added && m.noParentsPseudoID != "" {
			// Map to pseudo node
			id := MakePseudoNodeID(m.noParentsPseudoID, report.ExtractHostID(n))
			ret.addChildAndChildren(n, id, Pseudo)
		}
	}
	return ret.result(input)
}
