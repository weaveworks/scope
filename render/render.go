package render

import (
	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes
type Renderer interface {
	Render(report.Report) report.RenderableNodes
	EdgeMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers
type Reduce []Renderer

// Render produces a set of RenderableNodes given a Report
func (r Reduce) Render(rpt report.Report) report.RenderableNodes {
	result := report.RenderableNodes{}
	for _, renderer := range r {
		result.Merge(renderer.Render(rpt))
	}
	return result
}

// EdgeMetadata produces an AggregateMetadata for a given edge
func (r Reduce) EdgeMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	metadata := report.AggregateMetadata{}
	for _, renderer := range r {
		metadata.Merge(renderer.EdgeMetadata(rpt, localID, remoteID))
	}
	return metadata
}

// Map is a Renderer which produces a set of RendererNodes by using a
// Mapper functions and topology selector.
type Map struct {
	Selector report.TopologySelector
	Mapper   report.MapFunc
	Pseudo   report.PseudoFunc
}

// Render produces a set of RenderableNodes given a Report
func (m Map) Render(rpt report.Report) report.RenderableNodes {
	return m.Selector(rpt).RenderBy(m.Mapper, m.Pseudo)
}

// EdgeMetadata produces an AggregateMetadata for a given edge
func (m Map) EdgeMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return m.Selector(rpt).EdgeMetadata(m.Mapper, localID, remoteID).Transform()
}
