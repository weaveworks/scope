package render

import (
	"github.com/weaveworks/scope/report"
)

// ECSTaskRenderer is a Renderer for Amazon ECS tasks.
var ECSTaskRenderer = ConditionalRenderer(renderECSTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Container),
		MakeReduce(
			MakeMap(
				Map2Parent([]string{report.ECSTask}, NoParentsPseudo, UnmanagedID, nil),
				MakeFilter(
					IsRunning,
					ContainerWithImageNameRenderer,
				),
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
				Map2Parent([]string{report.ECSService}, NoParentsDrop, "", nil),
				ECSTaskRenderer,
			),
			SelectECSService,
		),
	),
)

func renderECSTopologies(rpt report.Report) bool {
	return len(rpt.ECSTask.Nodes)+len(rpt.ECSService.Nodes) >= 1
}
