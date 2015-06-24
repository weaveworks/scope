package render

import (
	"log"

	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes.
type Renderer interface {
	Render(report.Report) RenderableNodes
	AggregateMetadata(rpt report.Report, localID, remoteID string) AggregateMetadata
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers.
type Reduce []Renderer

// Map is a Renderer which produces a set of RenderableNodes from the set of
// RenderableNodes produced by another Renderer.
type Map struct {
	MapFunc
	Renderer
}

// LeafMap is a Renderer which produces a set of RenderableNodes from a report.Topology
// by using a map function and topology selector.
type LeafMap struct {
	Selector report.TopologySelector
	Mapper   LeafMapFunc
	Pseudo   PseudoFunc
}

// FilterUnconnected is a Renderer which filters out unconnected nodes.
type FilterUnconnected struct {
	Renderer
}

// MakeReduce is the only sane way to produce a Reduce Renderer.
func MakeReduce(renderers ...Renderer) Renderer {
	return Reduce(renderers)
}

// Render produces a set of RenderableNodes given a Report.
func (r Reduce) Render(rpt report.Report) RenderableNodes {
	result := RenderableNodes{}
	for _, renderer := range r {
		result.Merge(renderer.Render(rpt))
	}
	return result
}

// AggregateMetadata produces an AggregateMetadata for a given edge.
func (r Reduce) AggregateMetadata(rpt report.Report, localID, remoteID string) AggregateMetadata {
	metadata := AggregateMetadata{}
	for _, renderer := range r {
		metadata.Merge(renderer.AggregateMetadata(rpt, localID, remoteID))
	}
	return metadata
}

// Render transforms a set of RenderableNodes produces by another Renderer.
// using a map function
func (m Map) Render(rpt report.Report) RenderableNodes {
	output, _ := m.render(rpt)
	return output
}

func (m Map) render(rpt report.Report) (RenderableNodes, map[string]string) {
	input := m.Renderer.Render(rpt)
	output := RenderableNodes{}
	mapped := map[string]string{}             // input node ID -> output node ID
	adjacencies := map[string]report.IDList{} // output node ID -> input node Adjacencies

	for _, inRenderable := range input {
		outRenderable, ok := m.MapFunc(inRenderable)
		if !ok {
			continue
		}

		existing, ok := output[outRenderable.ID]
		if ok {
			outRenderable.Merge(existing)
		}

		output[outRenderable.ID] = outRenderable
		mapped[inRenderable.ID] = outRenderable.ID
		adjacencies[outRenderable.ID] = adjacencies[outRenderable.ID].Add(inRenderable.Adjacency...)
	}

	// Rewrite Adjacency for new node IDs.
	// NB we don't do pseudo nodes here; we assume the input graph
	// is properly-connected, and if the map func dropped a node,
	// we drop links to it.
	for outNodeID, inAdjacency := range adjacencies {
		outAdjacency := report.MakeIDList()
		for _, inAdjacent := range inAdjacency {
			if outAdjacent, ok := mapped[inAdjacent]; ok {
				outAdjacency = outAdjacency.Add(outAdjacent)
			}
		}
		outNode := output[outNodeID]
		outNode.Adjacency = outAdjacency
		output[outNodeID] = outNode
	}

	return output, mapped
}

// AggregateMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (m Map) AggregateMetadata(rpt report.Report, srcRenderableID, dstRenderableID string) AggregateMetadata {
	// First we need to map the ids in this layer into the ids in the underlying layer
	_, mapped := m.render(rpt)        // this maps from old -> new
	inverted := map[string][]string{} // this maps from new -> old(s)
	for k, v := range mapped {
		existing := inverted[v]
		existing = append(existing, k)
		inverted[v] = existing
	}

	// Now work out a slice of edges this edge is constructed from
	oldEdges := []struct{ src, dst string }{}
	for _, oldSrcID := range inverted[srcRenderableID] {
		for _, oldDstID := range inverted[dstRenderableID] {
			oldEdges = append(oldEdges, struct{ src, dst string }{oldSrcID, oldDstID})
		}
	}

	// Now recurse for each old edge
	output := AggregateMetadata{}
	for _, edge := range oldEdges {
		metadata := m.Renderer.AggregateMetadata(rpt, edge.src, edge.dst)
		output.Merge(metadata)
	}
	return output
}

// Render transforms a given Report into a set of RenderableNodes, which
// the UI will render collectively as a graph. Note that a RenderableNode will
// always be rendered with other nodes, and therefore contains limited detail.
//
// Nodes with the same mapped IDs will be merged.
func (m LeafMap) Render(rpt report.Report) RenderableNodes {
	var (
		t             = m.Selector(rpt)
		nodes         = RenderableNodes{}
		localNetworks = LocalNetworks(rpt)
	)

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	source2mapped := map[string]string{} // source node ID -> mapped node ID
	for nodeID, metadata := range t.NodeMetadatas {
		mapped, ok := m.Mapper(metadata)
		if !ok {
			continue
		}

		// mapped.ID needs not be unique over all addressIDs. If not, we merge with
		// the existing data, on the assumption that the MapFunc returns the same
		// data.
		existing, ok := nodes[mapped.ID]
		if ok {
			mapped.Merge(existing)
		}

		origins := mapped.Origins
		origins = origins.Add(nodeID)
		origins = origins.Add(metadata[report.HostNodeID])
		mapped.Origins = origins

		nodes[mapped.ID] = mapped
		source2mapped[nodeID] = mapped.ID
	}

	// Walk the graph and make connections.
	for src, dsts := range t.Adjacency {
		var (
			srcNodeID, ok     = report.ParseAdjacencyID(src)
			srcRenderableID   = source2mapped[srcNodeID] // must exist
			srcRenderableNode = nodes[srcRenderableID]   // must exist
		)
		if !ok {
			log.Printf("bad adjacency ID %q", src)
			continue
		}

		for _, dstNodeID := range dsts {
			dstRenderableID, ok := source2mapped[dstNodeID]
			if !ok {
				pseudoNode, ok := m.Pseudo(srcNodeID, srcRenderableNode, dstNodeID, localNetworks)
				if !ok {
					continue
				}
				dstRenderableID = pseudoNode.ID
				nodes[dstRenderableID] = pseudoNode
				source2mapped[dstNodeID] = dstRenderableID
			}

			srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstRenderableID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcNodeID)
			edgeID := report.MakeEdgeID(srcNodeID, dstNodeID)
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				srcRenderableNode.AggregateMetadata.Merge(AggregateMetadataOf(md))
			}
		}

		nodes[srcRenderableID] = srcRenderableNode
	}

	return nodes
}

// AggregateMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (m LeafMap) AggregateMetadata(rpt report.Report, srcRenderableID, dstRenderableID string) AggregateMetadata {
	t := m.Selector(rpt)
	metadata := report.EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		src, dst, ok := report.ParseEdgeID(edgeID)
		if !ok {
			log.Printf("bad edge ID %q", edgeID)
			continue
		}
		if src != report.TheInternet {
			mapped, _ := m.Mapper(t.NodeMetadatas[src])
			src = mapped.ID
		}
		if dst != report.TheInternet {
			mapped, _ := m.Mapper(t.NodeMetadatas[dst])
			dst = mapped.ID
		}
		if src == srcRenderableID && dst == dstRenderableID {
			metadata.Flatten(edgeMeta)
		}
	}
	return AggregateMetadataOf(metadata)
}

// Render produces a set of RenderableNodes given a Report
func (f FilterUnconnected) Render(rpt report.Report) RenderableNodes {
	return OnlyConnected(f.Renderer.Render(rpt))
}

// OnlyConnected filters out unconnected RenderedNodes
func OnlyConnected(input RenderableNodes) RenderableNodes {
	output := RenderableNodes{}
	for id, node := range input {
		if len(node.Adjacency) == 0 {
			continue
		}

		output[id] = node
		for _, id := range node.Adjacency {
			output[id] = input[id]
		}
	}
	return output
}
