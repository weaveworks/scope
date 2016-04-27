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

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = ApplyDecorators(
	MakeFilter(
		func(n report.Node) bool {
			// Drop deleted or empty pods
			state, ok := n.Latest.Lookup(kubernetes.PodState)
			return HasChildren(report.Container)(n) && (!ok || state != kubernetes.StateDeleted)
		},
		MakeReduce(
			MakeFilter(
				func(n report.Node) bool {
					// Drop unconnected pseudo nodes (could appear due to filtering)
					_, isConnected := n.Latest.Lookup(IsConnected)
					return n.Topology != Pseudo || isConnected
				},
				ColorConnected(MakeMap(
					MapContainer2Pod,
					ContainerWithImageNameRenderer,
				)),
			),
			SelectPod,
		),
	),
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = ApplyDecorators(
	FilterEmpty(report.Pod,
		MakeReduce(
			MakeMap(
				MapPod2Service,
				PodRenderer,
			),
			SelectService,
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

// MapPod2Service maps pod Nodes to service Nodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a pod graph to get that info.
func MapPod2Service(pod report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if pod.Topology == Pseudo {
		return report.Nodes{pod.ID: pod}
	}

	// Otherwise, if some some reason the pod doesn't have a service_ids (maybe
	// slightly out of sync reports, or its not in a service), just drop it
	namespace, ok := pod.Latest.Lookup(kubernetes.Namespace)
	if !ok {
		return report.Nodes{}
	}
	serviceIDs, ok := pod.Sets.Lookup(kubernetes.ServiceIDs)
	if !ok {
		return report.Nodes{}
	}

	result := report.Nodes{}
	for _, serviceID := range serviceIDs {
		serviceName := strings.TrimPrefix(serviceID, namespace+"/")
		id := report.MakeServiceNodeID(namespace, serviceName)
		node := NewDerivedNode(id, pod).WithTopology(report.Service)
		node.Counters = node.Counters.Add(pod.Topology, 1)
		result[id] = node
	}
	return result
}
