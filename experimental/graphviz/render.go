package main

import (
	"fmt"

	"$GITHUB_URI/render"
	"$GITHUB_URI/render/detailed"
	"$GITHUB_URI/report"
)

func renderTo(rpt report.Report, topology string) (detailed.NodeSummaries, error) {
	renderer, ok := map[string]render.Renderer{
		"processes":           render.FilterUnconnected(render.ProcessWithContainerNameRenderer),
		"processes-by-name":   render.FilterUnconnected(render.ProcessNameRenderer),
		"containers":          render.ContainerWithImageNameRenderer,
		"containers-by-image": render.ContainerImageRenderer,
		"hosts":               render.HostRenderer,
	}[topology]
	if !ok {
		return detailed.NodeSummaries{}, fmt.Errorf("unknown topology %v", topology)
	}
	return detailed.Summaries(rpt, renderer.Render(rpt, render.FilterNoop)), nil
}
