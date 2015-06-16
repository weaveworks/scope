package render

import (
	"log"

	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes
type Renderer interface {
	Render(report.Report) RenderableNodes
	AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata
}

// Reduce renderer is a Renderer which merges together the output of several
// other renderers
type Reduce []Renderer

// Render produces a set of RenderableNodes given a Report
func (r Reduce) Render(rpt report.Report) RenderableNodes {
	result := RenderableNodes{}
	for _, renderer := range r {
		result.Merge(renderer.Render(rpt))
	}
	return result
}

// AggregateMetadata produces an AggregateMetadata for a given edge
func (r Reduce) AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	metadata := report.AggregateMetadata{}
	for _, renderer := range r {
		metadata.Merge(renderer.AggregateMetadata(rpt, localID, remoteID))
	}
	return metadata
}

// Map is a Renderer which produces a set of RendererNodes by using a
// Mapper functions and topology selector.
type Map struct {
	Selector report.TopologySelector
	Mapper   MapFunc
	Pseudo   PseudoFunc
}

// Render produces a set of RenderableNodes given a Report
func (m Map) Render(rpt report.Report) RenderableNodes {
	return Topology(m.Selector(rpt), m.Mapper, m.Pseudo)
}

// Topology transforms a given Topology into a set of RenderableNodes, which
// the UI will render collectively as a graph. Note that a RenderableNode will
// always be rendered with other nodes, and therefore contains limited detail.
//
// RenderBy takes a a MapFunc, which defines how to group and label nodes. Npdes
// with the same mapped IDs will be merged.
func Topology(t report.Topology, mapFunc MapFunc, pseudoFunc PseudoFunc) RenderableNodes {
	nodes := RenderableNodes{}

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	var (
		source2mapped = map[string]string{} // source node ID -> mapped node ID
		source2host   = map[string]string{} // source node ID -> origin host ID
	)
	for nodeID, metadata := range t.NodeMetadatas {
		mapped, ok := mapFunc(metadata)
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

		mapped.Origins = mapped.Origins.Add(nodeID)
		nodes[mapped.ID] = mapped
		source2mapped[nodeID] = mapped.ID
		source2host[nodeID] = metadata[report.HostNodeID]
	}

	// Walk the graph and make connections.
	for src, dsts := range t.Adjacency {
		var (
			srcNodeID, ok = report.ParseAdjacencyID(src)
			//srcOriginHostID, _, ok2 = ParseNodeID(srcNodeID)
			srcHostNodeID     = source2host[srcNodeID]
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
				pseudoNode, ok := pseudoFunc(srcNodeID, srcRenderableNode, dstNodeID)
				if !ok {
					continue
				}
				dstRenderableID = pseudoNode.ID
				nodes[dstRenderableID] = pseudoNode
				source2mapped[dstNodeID] = dstRenderableID
			}

			srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstRenderableID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcHostNodeID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcNodeID)
			edgeID := report.MakeEdgeID(srcNodeID, dstNodeID)
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				srcRenderableNode.Metadata.Merge(md.Transform())
			}
		}

		nodes[srcRenderableID] = srcRenderableNode
	}

	return nodes
}

// AggregateMetadata produces an AggregateMetadata for a given edge
func (m Map) AggregateMetadata(rpt report.Report, localID, remoteID string) report.AggregateMetadata {
	return edgeMetadata(m.Selector(rpt), m.Mapper, localID, remoteID).Transform()
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func edgeMetadata(t report.Topology, mapFunc MapFunc, srcRenderableID, dstRenderableID string) report.EdgeMetadata {
	metadata := report.EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		src, dst, ok := report.ParseEdgeID(edgeID)
		if !ok {
			log.Printf("bad edge ID %q", edgeID)
			continue
		}
		if src != report.TheInternet {
			mapped, _ := mapFunc(t.NodeMetadatas[src])
			src = mapped.ID
		}
		if dst != report.TheInternet {
			mapped, _ := mapFunc(t.NodeMetadatas[dst])
			dst = mapped.ID
		}
		if src == srcRenderableID && dst == dstRenderableID {
			metadata.Flatten(edgeMeta)
		}
	}
	return metadata
}
