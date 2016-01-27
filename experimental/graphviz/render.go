package main

import (
	"fmt"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

func renderTo(rpt report.Report, topology string) (render.RenderableNodes, error) {
	renderer, ok := map[string]render.Renderer{
		"applications":         render.FilterUnconnected(render.ProcessWithContainerNameRenderer),
		"applications-by-name": render.FilterUnconnected(render.ProcessNameRenderer),
		"containers":           render.ContainerWithImageNameRenderer,
		"containers-by-image":  render.ContainerImageRenderer,
		"hosts":                render.HostRenderer,
	}[topology]
	if !ok {
		return render.MakeRenderableNodes(), fmt.Errorf("unknown topology %v", topology)
	}
	return renderer.Render(rpt), nil
}
