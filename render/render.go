package render

import (
	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes.
type Renderer interface {
	Render(report.Report) RenderableNodes
	EdgeMetadata(rpt report.Report, localID, remoteID string) report.EdgeMetadata
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
	return Reduce(renderers)
}

// Render produces a set of RenderableNodes given a Report.
func (r Reduce) Render(rpt report.Report) RenderableNodes {
	result := RenderableNodes{}
	for _, renderer := range r {
		result = result.Merge(renderer.Render(rpt))
	}
	return result
}

// EdgeMetadata produces an EdgeMetadata for a given edge.
func (r Reduce) EdgeMetadata(rpt report.Report, localID, remoteID string) report.EdgeMetadata {
	metadata := report.EdgeMetadata{}
	for _, renderer := range r {
		metadata = metadata.Merge(renderer.EdgeMetadata(rpt, localID, remoteID))
	}
	return metadata
}

// Stats implements Renderer
func (r Reduce) Stats(rpt report.Report) Stats {
	var result Stats
	for _, renderer := range r {
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

// Render transforms a set of RenderableNodes produces by another Renderer.
// using a map function
func (m Map) Render(rpt report.Report) RenderableNodes {
	output, _ := m.render(rpt)
	return output
}

// Stats implements Renderer
func (m Map) Stats(rpt report.Report) Stats {
	// There doesn't seem to be an instance where we want stats to recurse
	// through Maps - for instance we don't want to see the number of filtered
	// processes in the container renderer.
	return Stats{}
}

func (m Map) render(rpt report.Report) (RenderableNodes, map[string]report.IDList) {
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

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (m Map) EdgeMetadata(rpt report.Report, srcRenderableID, dstRenderableID string) report.EdgeMetadata {
	// First we need to map the ids in this layer into the ids in the underlying layer
	_, mapped := m.render(rpt)        // this maps from old -> new
	inverted := map[string][]string{} // this maps from new -> old(s)
	for k, vs := range mapped {
		for _, v := range vs {
			existing := inverted[v]
			existing = append(existing, k)
			inverted[v] = existing
		}
	}

	// Now work out a slice of edges this edge is constructed from
	oldEdges := []struct{ src, dst string }{}
	for _, oldSrcID := range inverted[srcRenderableID] {
		for _, oldDstID := range inverted[dstRenderableID] {
			oldEdges = append(oldEdges, struct{ src, dst string }{oldSrcID, oldDstID})
		}
	}

	// Now recurse for each old edge
	output := report.EdgeMetadata{}
	for _, edge := range oldEdges {
		metadata := m.Renderer.EdgeMetadata(rpt, edge.src, edge.dst)
		output = output.Merge(metadata)
	}
	return output
}
