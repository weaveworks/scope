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

func renderKubernetesTopologies(rpt report.Report) bool {
	return len(rpt.Pod.Nodes)+len(rpt.Service.Nodes)+len(rpt.Deployment.Nodes)+len(rpt.ReplicaSet.Nodes) >= 1
}

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(MakeFilter(
		func(n report.Node) bool {
			state, ok := n.Latest.Lookup(kubernetes.State)
			return (!ok || state != kubernetes.StateDeleted)
		},
		MakeMap(
			PropagateSingleMetrics(report.Container),
			MakeReduce(
				MakeMap(
					MapContainer2Pod,
					ContainerWithImageNameRenderer,
				),
				ShortLivedConnectionJoin(SelectPod, MapPod2IP),
				SelectPod,
			),
		),
	)),
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(
		MakeMap(
			PropagateSingleMetrics(report.Pod),
			MakeReduce(
				MakeMap(
					Map2Service,
					PodRenderer,
				),
				SelectService,
			),
		),
	),
)

// DeploymentRenderer is a Renderer which produces a renderable kubernetes deployments
// graph by merging the pods graph and the deployments topology.
var DeploymentRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(
		MakeMap(
			PropagateSingleMetrics(report.ReplicaSet),
			MakeReduce(
				MakeMap(
					Map2Deployment,
					ReplicaSetRenderer,
				),
				SelectDeployment,
			),
		),
	),
)

// ReplicaSetRenderer is a Renderer which produces a renderable kubernetes replica sets
// graph by merging the pods graph and the replica sets topology.
var ReplicaSetRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(
		MakeMap(
			PropagateSingleMetrics(report.Pod),
			MakeReduce(
				MakeMap(
					Map2ReplicaSet,
					PodRenderer,
				),
				SelectReplicaSet,
			),
		),
	),
)

// MapContainer2Pod maps container Nodes to pod
// Nodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Unmanaged"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2Pod(n report.Node, _ report.Networks) report.Nodes {
	// Uncontained becomes unmanaged in the pods view
	if strings.HasPrefix(n.ID, MakePseudoNodeID(UncontainedID)) {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}

	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	// Ignore non-running containers
	if state, ok := n.Latest.Lookup(docker.ContainerState); ok && state != docker.StateRunning {
		return report.Nodes{}
	}

	// Otherwise, if some some reason the container doesn't have a pod uid (maybe
	// slightly out of sync reports, or its not in a pod), make it part of unmanaged.
	uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
	if !ok {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}

	id := report.MakePodNodeID(uid)
	node := NewDerivedNode(id, n).
		WithTopology(report.Pod)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
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

// The various ways of grouping pods
var (
	Map2Service    = Map2Parent(report.Service)
	Map2Deployment = Map2Parent(report.Deployment)
	Map2ReplicaSet = Map2Parent(report.ReplicaSet)
)

// Map2Parent maps Nodes to some parent grouping.
func Map2Parent(topology string) func(n report.Node, _ report.Networks) report.Nodes {
	return func(n report.Node, _ report.Networks) report.Nodes {
		// Propagate all pseudo nodes
		if n.Topology == Pseudo {
			return report.Nodes{n.ID: n}
		}

		// Otherwise, if some some reason the node doesn't have any of these ids
		// (maybe slightly out of sync reports, or its not in this group), just
		// drop it
		groupIDs, ok := n.Parents.Lookup(topology)
		if !ok {
			return report.Nodes{}
		}

		result := report.Nodes{}
		for _, id := range groupIDs {
			node := NewDerivedNode(id, n).WithTopology(topology)
			node.Counters = node.Counters.Add(n.Topology, 1)

			// When mapping replica(tionController)s(ets) to deployments
			// we must propagate the pod counter.
			if n.Topology != report.Pod {
				if count, ok := n.Counters.Lookup(report.Pod); ok {
					node.Counters = node.Counters.Add(report.Pod, count)
				}
			}

			result[id] = node
		}
		return result
	}
}
