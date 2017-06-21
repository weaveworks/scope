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
					Map2Parent(report.Pod, UnmanagedID, nil),
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
				Map2Parent(report.Service, "", nil),
				PodRenderer,
			),
			SelectService,
		),
	),
)

// DeploymentRenderer is a Renderer which produces a renderable kubernetes deployments
// graph by merging the pods graph and the deployments topology.
var DeploymentRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeMap(
		PropagateSingleMetrics(report.ReplicaSet),
		MakeReduce(
			MakeMap(
				Map2Parent(report.Deployment, "", mapPodCounts),
				ReplicaSetRenderer,
			),
			SelectDeployment,
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
				Map2Parent(report.ReplicaSet, "", nil),
				PodRenderer,
			),
			SelectReplicaSet,
		),
	),
)

// DaemonSetRenderer is a Renderer which produces a renderable kubernetes daemonsets
// graph by merging the pods graph and the daemonsets topology.
var DaemonSetRenderer = ConditionalRenderer(renderKubernetesTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Pod),
		MakeReduce(
			MakeMap(
				Map2Parent(report.DaemonSet, "", nil),
				PodRenderer,
			),
			SelectDaemonSet,
		),
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
	// The topology ID of the parents
	topology string,
	// Either the ID prefix of the pseudo node to use for nodes without
	// any parents in the group, eg. UnmanagedID, or "" to drop nodes without any parents.
	noParentsPseudoID string,
	// Optional (can be nil) function to modify any parent nodes,
	// eg. to copy over details from the original node.
	modifyMappedNode func(parent, original report.Node) report.Node,
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

		// If some some reason the node doesn't have any of these ids
		// (maybe slightly out of sync reports, or its not in this group),
		// either drop it or put it in Uncontained/Unmanaged/whatever if one was given
		groupIDs, ok := n.Parents.Lookup(topology)
		if !ok || len(groupIDs) == 0 {
			if noParentsPseudoID == "" {
				return report.Nodes{}
			}
			id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
			node := NewDerivedPseudoNode(id, n)
			return report.Nodes{id: node}
		}

		result := report.Nodes{}
		for _, id := range groupIDs {
			node := NewDerivedNode(id, n).WithTopology(topology)
			node.Counters = node.Counters.Add(n.Topology, 1)
			if modifyMappedNode != nil {
				node = modifyMappedNode(node, n)
			}
			result[id] = node
		}
		return result
	}
}
