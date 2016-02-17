package render

import (
	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes.
type Renderer interface {
	Render(report.Report) RenderableNodes
	Stats(report.Report) Stats
}

// Stats is the type returned by Renderer.Stats
type Stats struct {
	FilteredNodes int
}

func (s Stats) merge(other Stats) Stats {
	return Stats{
		FilteredNodes: s.FilteredNodes + other.FilteredNodes,
	}
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers.
type Reduce []Renderer

// MakeReduce is the only sane way to produce a Reduce Renderer.
func MakeReduce(renderers ...Renderer) Renderer {
	r := Reduce(renderers)
	return Memoise(&r)
}

// Render produces a set of RenderableNodes given a Report.
func (r *Reduce) Render(rpt report.Report) RenderableNodes {
	result := RenderableNodes{}
	for _, renderer := range *r {
		result = result.Merge(renderer.Render(rpt))
	}
	return result
}

// Stats implements Renderer
func (r *Reduce) Stats(rpt report.Report) Stats {
	var result Stats
	for _, renderer := range *r {
		result = result.merge(renderer.Stats(rpt))
	}
	return result
}

// Map is a Renderer which produces a set of RenderableNodes from the set of
// RenderableNodes produced by another Renderer.
type Map struct {
	MapFunc
	Renderer
}

// MakeMap makes a new Map
func MakeMap(f MapFunc, r Renderer) Renderer {
	return Memoise(&Map{f, r})
}

// Render transforms a set of RenderableNodes produces by another Renderer.
// using a map function
func (m *Map) Render(rpt report.Report) RenderableNodes {
	output, _ := m.render(rpt)
	return output
}

// Stats implements Renderer
func (m *Map) Stats(rpt report.Report) Stats {
	// There doesn't seem to be an instance where we want stats to recurse
	// through Maps - for instance we don't want to see the number of filtered
	// processes in the container renderer.
	return Stats{}
}

func (m *Map) render(rpt report.Report) (RenderableNodes, map[string]report.IDList) {
	var (
		input         = m.Renderer.Render(rpt)
		output        = RenderableNodes{}
		mapped        = map[string]report.IDList{} // input node ID -> output node IDs
		adjacencies   = map[string]report.IDList{} // output node ID -> input node Adjacencies
		localNetworks = LocalNetworks(rpt)
	)

	// Rewrite all the nodes according to the map function
	for _, inRenderable := range input {
		for _, outRenderable := range m.MapFunc(inRenderable, localNetworks) {
			existing, ok := output[outRenderable.ID]
			if ok {
				outRenderable = outRenderable.Merge(existing)
			}

			output[outRenderable.ID] = outRenderable
			mapped[inRenderable.ID] = mapped[inRenderable.ID].Add(outRenderable.ID)
			adjacencies[outRenderable.ID] = adjacencies[outRenderable.ID].Merge(inRenderable.Adjacency)
		}
	}

	// Rewrite Adjacency for new node IDs.
	for outNodeID, inAdjacency := range adjacencies {
		outAdjacency := report.MakeIDList()
		for _, inAdjacent := range inAdjacency {
			for _, outAdjacent := range mapped[inAdjacent] {
				outAdjacency = outAdjacency.Add(outAdjacent)
			}
		}
		outNode := output[outNodeID]
		outNode.Adjacency = outAdjacency
		output[outNodeID] = outNode
	}

	return output, mapped
}
