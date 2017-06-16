package render

import (
	"github.com/weaveworks/scope/report"
)

// SwarmServiceRenderer is a Renderer for Docker Swarm services
var SwarmServiceRenderer = ConditionalRenderer(renderSwarmTopologies,
	renderParents(
		report.Container, []string{report.SwarmService}, NoParentsPseudo, UnmanagedID, nil,
		MakeFilter(
			IsRunning,
			ContainerWithImageNameRenderer,
		),
	),
)

func renderSwarmTopologies(rpt report.Report) bool {
	return len(rpt.SwarmService.Nodes) >= 1
}
