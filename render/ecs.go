package render

import (
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// ECSTaskRenderer is a Renderer for Amazon ECS tasks.
var ECSTaskRenderer = ConditionalRenderer(renderECSTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Container),
		MakeReduce(
			MakeMap(
				MapContainer2ECSTask,
				ContainerWithImageNameRenderer,
			),
			SelectECSTask,
		),
	),
)

// ECSServiceRenderer is a Renderer for Amazon ECS services.
var ECSServiceRenderer = ConditionalRenderer(renderECSTopologies,
	MakeMap(
		PropagateSingleMetrics(report.ECSTask),
		MakeReduce(
			MakeMap(
				Map2Parent(report.ECSService),
				ECSTaskRenderer,
			),
			SelectECSService,
		),
	),
)

// MapContainer2ECSTask maps container Nodes to ECS Task
// Nodes.
//
// If this function is given a node without an ECS Task parent
// (including other pseudo nodes), it will produce an "Unmanaged"
// pseudo node.
//
// TODO: worth merging with MapContainer2Pod?
func MapContainer2ECSTask(n report.Node, _ report.Networks) report.Nodes {
	// Uncontained becomes unmanaged in the tasks view
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

	taskIDSet, ok := n.Parents.Lookup(report.ECSTask)
	if !ok || len(taskIDSet) == 0 {
		id := MakePseudoNodeID(UnmanagedID, report.ExtractHostID(n))
		node := NewDerivedPseudoNode(id, n)
		return report.Nodes{id: node}
	}
	nodeID := taskIDSet[0]
	node := NewDerivedNode(nodeID, n).WithTopology(report.ECSTask)
	// Propagate parent service
	if serviceIDSet, ok := n.Parents.Lookup(report.ECSService); ok {
		node = node.WithParents(report.MakeSets().Add(report.ECSService, serviceIDSet))
	}
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{nodeID: node}
}

func renderECSTopologies(rpt report.Report) bool {
	return len(rpt.ECSTask.Nodes)+len(rpt.ECSService.Nodes) >= 1
}
