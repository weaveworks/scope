package render

import (
	"strings"

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
func PodRenderer(filter Decorator) Renderer {
	return FilterEmpty(report.Container,
		MakeReduce(
			MakeSilentFilter(
				func(n report.Node) bool {
					// Drop unconnected pseudo nodes (could appear due to filtering)
					_, isConnected := n.Latest.Lookup(IsConnected)
					return n.Topology != Pseudo || isConnected
				},
				ColorConnected(MakeMap(
					MapContainer2Pod,
					filter(ContainerWithImageNameRenderer),
				)),
			),
			SelectPod,
		),
	)
}

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
func PodServiceRenderer(filter Decorator) Renderer {
	return FilterEmpty(report.Pod,
		MakeReduce(
			MakeMap(
				MapPod2Service,
				PodRenderer(filter),
			),
			SelectService,
		),
	)
}

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
	// Uncontainerd becomes unmanaged in the pods view
	if strings.HasPrefix(n.ID, MakePseudoNodeID(UncontainedID)) {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}

	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a pod_id (maybe
	// slightly out of sync reports, or its not in a pod), just drop it
	namespace, ok := n.Latest.Lookup(kubernetes.Namespace)
	if !ok {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}
	podID, ok := n.Latest.Lookup(kubernetes.PodID)
	if !ok {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}
	podName := strings.TrimPrefix(podID, namespace+"/")
	id := report.MakePodNodeID(namespace, podName)

	// Due to a bug in kubernetes, addon pods on the master node are not returned
	// from the API. Adding the namespace and pod name is a workaround until
	// https://github.com/kubernetes/kubernetes/issues/14738 is fixed.
	return report.Nodes{
		id: NewDerivedNode(id, n).
			WithTopology(report.Pod).
			WithLatests(map[string]string{
				kubernetes.Namespace: namespace,
				kubernetes.PodName:   podName,
			}),
	}
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
	ids, ok := pod.Latest.Lookup(kubernetes.ServiceIDs)
	if !ok {
		return report.Nodes{}
	}

	result := report.Nodes{}
	for _, serviceID := range strings.Fields(ids) {
		serviceName := strings.TrimPrefix(serviceID, namespace+"/")
		id := report.MakeServiceNodeID(namespace, serviceName)
		node := NewDerivedNode(id, pod).WithTopology(report.Service)
		node.Counters = node.Counters.Add(pod.Topology, 1)
		result[id] = node
	}
	return result
}
