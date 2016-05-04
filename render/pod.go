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
	return len(rpt.Pod.Nodes)+len(rpt.Service.Nodes) > 1
}

// PodRenderer is a Renderer which produces a renderable kubernetes
// graph by merging the container graph and the pods topology.
var PodRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(MakeFilter(
		func(n report.Node) bool {
			state, ok := n.Latest.Lookup(kubernetes.PodState)
			return (!ok || state != kubernetes.StateDeleted)
		},
		MakeReduce(
			MakeMap(
				MapContainer2Pod,
				ContainerWithImageNameRenderer,
			),
			SelectPod,
		),
	)),
)

// PodServiceRenderer is a Renderer which produces a renderable kubernetes services
// graph by merging the pods graph and the services topology.
var PodServiceRenderer = ConditionalRenderer(renderKubernetesTopologies,
	ApplyDecorators(FilterEmpty(report.Pod,
		MakeReduce(
			MakeMap(
				MapPod2Service,
				PodRenderer,
			),
			SelectService,
		),
	)),
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

// The various ways of grouping pods
var (
	MapPod2Service    = MapPod2Grouping(report.Service, kubernetes.ServiceIDs, report.MakeServiceNodeID)
)

// MapPod2Grouping maps pod Nodes to some kubernetes grouping.
func MapPod2Grouping(topology, setKey string, idMaker func(uid string) string) func(pod report.Node, _ report.Networks) report.Nodes {
	return func(pod report.Node, _ report.Networks) report.Nodes {
		// Propagate all pseudo nodes
		if pod.Topology == Pseudo {
			return report.Nodes{pod.ID: pod}
		}

		// Otherwise, if some some reason the pod doesn't have any of these ids
		// (maybe slightly out of sync reports, or its not in this group), just
		// drop it
		groupIDs, ok := pod.Sets.Lookup(setKey)
		if !ok {
			return report.Nodes{}
		}

		result := report.Nodes{}
		for _, groupID := range groupIDs {
			id := idMaker(groupID)
			node := NewDerivedNode(id, pod).WithTopology(topology)
			node.Counters = node.Counters.Add(pod.Topology, 1)
			result[id] = node
		}
		return result
	}
}
