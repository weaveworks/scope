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
	return len(rpt.Pod.Nodes)+len(rpt.Service.Nodes)+len(rpt.Deployment.Nodes)+len(rpt.ReplicaSet.Nodes)+len(rpt.DaemonSet.Nodes) >= 1
}

func isPauseContainer(n report.Node) bool {
	image, ok := n.Latest.Lookup(docker.ImageName)
	return ok && kubernetes.IsPauseImageName(image)
}

type noParentsActionEnum int

// Constants for specifying noParentsAction in Map2Parent
const (
	NoParentsPseudo noParentsActionEnum = iota
	NoParentsDrop
	NoParentsKeep
)

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeFilter(
		func(n report.Node) bool {
			state, ok := n.Latest.Lookup(kubernetes.State)
			return (!ok || state != kubernetes.StateDeleted)
		},
		MakeMap(
			PropagateSingleMetrics(report.Container),
			MakeReduce(
				MakeMap(
					Map2Parent([]string{report.Pod}, NoParentsPseudo, UnmanagedID, nil),
					MakeFilter(
						ComposeFilterFuncs(
							IsRunning,
							Complement(isPauseContainer),
						),
						ContainerWithImageNameRenderer,
					),
				),
				ConnectionJoin(SelectPod, MapPod2IP),
				SelectPod,
			),
		),
	),
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Pod),
		MakeReduce(
			MakeMap(
				Map2Parent([]string{report.Service}, NoParentsDrop, "", nil),
				PodRenderer,
			),
			SelectService,
		),
	),
)

// ReplicaSetRenderer is a Renderer which produces a renderable kubernetes replica sets
// graph by merging the pods graph and the replica sets topology.
var ReplicaSetRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Pod),
		MakeReduce(
			MakeMap(
				Map2Parent([]string{report.ReplicaSet}, NoParentsDrop, "", nil),
				PodRenderer,
			),
			SelectReplicaSet,
		),
	),
)

// KubeControllerRenderer is a Renderer which combines all the 'controller' topologies.
// We first map pods to daemonsets and replica sets, then to deployments since replica sets
// can map to those (the rest are passed through unchanged).
// Pods with no controller are mapped to 'Unmanaged'
// We can't simply combine the rendered graphs of the high level objects as they would never
// have connections to each other.
var KubeControllerRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeReduce(
		MakeFilter(
			Complement(IsTopology(report.ReplicaSet)),
			MakeMap(
				Map2Parent([]string{report.Deployment}, NoParentsKeep, "", mapPodCounts),
				MakeReduce(
					MakeMap(
						Map2Parent([]string{
							report.ReplicaSet,
							report.DaemonSet,
						}, NoParentsPseudo, UnmanagedID, nil),
						PodRenderer,
					),
					SelectReplicaSet,
					SelectDaemonSet,
				),
			),
		),
		SelectDeployment,
	),
)

func mapPodCounts(parent, original report.Node) report.Node {
	// When mapping ReplicaSets to Deployments, we want to propagate the Pods counter
	if count, ok := original.Counters.Lookup(report.Pod); ok {
		parent.Counters = parent.Counters.Add(report.Pod, count)
	}
	return parent
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
	// Choose what to do in the case of nodes with no parents. One of:
	//     NoParentsPseudo: Map them to a common pseudo node id with prefix noParentsPseudoID
	//     NoParentsDrop: Map them to no node.
	//     NoParentsKeep: Map them to themselves, preserving them in the new graph.
	noParentsAction noParentsActionEnum,
	// The ID prefix of the pseudo node to use for nodes without any parents in the group
	// if noParentsAction == Pseudo, eg. UnmanagedID
	noParentsPseudoID string,
	// Optional (can be nil) function to modify any parent nodes,
	// eg. to copy over details from the original node.
	modifyMappedNode func(parent, original report.Node) report.Node,
) MapFunc {
	return func(n report.Node, _ report.Networks) report.Nodes {
		// Uncontained becomes Unmanaged/whatever if noParentsAction == Pseudo
		if noParentsAction == NoParentsPseudo && strings.HasPrefix(n.ID, UncontainedIDPrefix) {
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
					if modifyMappedNode != nil {
						node = modifyMappedNode(node, n)
					}
					result[id] = node
				}
			}
		}

		if len(result) == 0 {
			switch noParentsAction {
			case NoParentsPseudo:
				// Map to pseudo node
				id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
				node := NewDerivedPseudoNode(id, n)
				result[id] = node
			case NoParentsKeep:
				// Pass n to output unmodified
				result[n.ID] = n
			case NoParentsDrop:
				// Do nothing, we will return an empty result
			}
		}

		return result
	}
}
