package render

import (
	"github.com/weaveworks/scope/report"
)

// ECSTaskRenderer is a Renderer for Amazon ECS tasks.
var ECSTaskRenderer = Memoise(ConditionalRenderer(renderECSTopologies,
	renderParents(
		report.Container, []string{report.ECSTask}, UnmanagedID,
		MakeFilter(
			IsRunning,
			ContainerWithImageNameRenderer,
		),
	),
))

// ECSServiceRenderer is a Renderer for Amazon ECS services.
//
// not memoised
var ECSServiceRenderer = ConditionalRenderer(renderECSTopologies,
	renderParents(
		report.ECSTask, []string{report.ECSService}, "",
		ECSTaskRenderer,
	),
)

func renderECSTopologies(rpt report.Report) bool {
	return len(rpt.ECSTask.Nodes)+len(rpt.ECSService.Nodes) >= 1
}
