package render

import (
	"fmt"
	"strings"

	"github.com/weaveworks/scope/common/mtime"
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
	ApplyDecorators(
		MakeFilter(
			func(n report.Node) bool {
				state, ok := n.Latest.Lookup(kubernetes.State)
				return (!ok || state != kubernetes.StateDeleted)
			},
			MakeMap(
				PropagateSingleMetrics(report.Container),
				MakeReduce(
					// 1st take pseudo nodes from the container renders
					MakeMap(
						MapContainer2PodPseudos,
						ContainerWithImageNameRenderer,
					),

					// 2nd take containers in pods on k8s >1.2
					MakeMap(
						MapContainer2PodUID,
						ContainerWithImageNameRenderer,
					),

					// 3rd take containers in pods on k8s <=1.1
					MakeMap(
						MapPodName2PodUID,
						MakeReduce(
							MakeMap(
								MapContainer2PodName,
								ContainerWithImageNameRenderer,
							),
							MakeMap(
								MapPodUID2PodName,
								SelectPod,
							),
						),
					),

					// Finally, take the actual pods
					SelectPod,
				),
			),
		),
	),
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

// MapContainer2PodPseudos propagates pseudo nodes from the Container topology.
func MapContainer2PodPseudos(n report.Node, _ report.Networks) report.Nodes {
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

	// For k8s >=1.2, we join based on uid - ignore these
	if _, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid"); ok {
		return report.Nodes{}
	}

	// For k8s <1.2, try and join by pod name - ignore these
	if _, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.name"); ok {
		return report.Nodes{}
	}

	// Otherwise, if some some reason the container doesn't have a pod uid (maybe
	// slightly out of sync reports, or its not in a pod), make it part of unmanaged.
	id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
	node := NewDerivedPseudoNode(id, n)
	return report.Nodes{id: node}
}

// MapContainer2PodUID maps container nodes to pods based on the pod uuid label,
// which only appears in k8s 1.2.
func MapContainer2PodUID(n report.Node, _ report.Networks) report.Nodes {
	// Ignore all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{}
	}

	// Ignore non-running containers
	if state, ok := n.Latest.Lookup(docker.ContainerState); ok && state != docker.StateRunning {
		return report.Nodes{}
	}

	// For k8s >=1.2, we join based on uid
	uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
	if !ok {
		return report.Nodes{}
	}

	id := report.MakePodNodeID(uid)
	node := NewDerivedNode(id, n).
		WithTopology(report.Pod)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}

// MapContainer2PodName maps container nodes to pods based on the pod name label,
// which only appears in k8s <1.2.  We don't do this all the time as it is less precise.
func MapContainer2PodName(n report.Node, _ report.Networks) report.Nodes {
	// Ignore all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{}
	}

	// Ignore non-running containers
	if state, ok := n.Latest.Lookup(docker.ContainerState); ok && state != docker.StateRunning {
		return report.Nodes{}
	}

	// Ignore container with a pod uid label
	uid, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.uid")
	if ok {
		return report.Nodes{}
	}

	// For k8s <1.2, try and join by pod name, which needs some gymnastics
	podName, ok := n.Latest.Lookup(docker.LabelPrefix + "io.kubernetes.pod.name")
	if !ok {
		return report.Nodes{}
	}

	id := report.MakePodNodeID(podName)
	node := NewDerivedNode(id, n).
		WithTopology(report.Pod)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}

const originalID = "original_id"

// MapPodUID2PodName maps pods-by-uuid to pod-by-name, for joining with
// containers from MapContainer2PodName.
func MapPodUID2PodName(n report.Node, _ report.Networks) report.Nodes {
	namespace, ok := n.Latest.Lookup(kubernetes.Namespace)
	if !ok {
		return report.Nodes{}
	}

	name, ok := n.Latest.Lookup(kubernetes.Name)
	if !ok {
		return report.Nodes{}
	}

	id := report.MakePodNodeID(fmt.Sprintf("%s/%s", namespace, name))
	node := NewDerivedNode(id, n).
		WithTopology(report.Pod).
		WithLatest(originalID, mtime.Now(), n.ID)
	node.Counters = node.Counters.Add(report.Pod, 1)
	return report.Nodes{id: node}
}

// MapPodName2PodUID maps pod-by-name back to pods-by-uuid
func MapPodName2PodUID(n report.Node, _ report.Networks) report.Nodes {
	// If there were more than one pod with same name (ie it got
	// deleted and recreated real quick) then ignore it, as container
	// join will be non-deterministic.
	if c, ok := n.Counters.Lookup(report.Pod); !ok || c > 1 {
		return report.Nodes{}
	}

	origID, ok := n.Latest.Lookup(originalID)
	if !ok {
		return report.Nodes{}
	}

	node := NewDerivedNode(origID, n).
		WithTopology(report.Pod)
	// replace children to exclude n, as it should
	// never appear in the output.
	node.Children = n.Children
	return report.Nodes{node.ID: node}
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
