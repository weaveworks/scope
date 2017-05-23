package render

import (
	"github.com/weaveworks/scope/report"
)

// SwarmServiceRenderer is a Renderer for Docker Swarm services
var SwarmServiceRenderer = ConditionalRenderer(renderSwarmTopologies,
	MakeMap(
		PropagateSingleMetrics(report.Container),
		MakeReduce(
			MakeMap(
				Map2Parent([]string{report.SwarmService}, NoParentsPseudo, UnmanagedID, nil),
				MakeFilter(
					IsRunning,
					ContainerWithImageNameRenderer,
				),
			),
			SelectSwarmService,
		),
	),
)

func renderSwarmTopologies(rpt report.Report) bool {
	return len(rpt.SwarmService.Nodes) >= 1
}
