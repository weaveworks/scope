package render

import (
	"github.com/weaveworks/scope/report"
)

// Renderer is something that can render a report to a set of RenderableNodes.
type Renderer interface {
	Render(report.Report) RenderableNodes
	EdgeMetadata(rpt report.Report, localID, remoteID string) report.EdgeMetadata
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
		result.Merge(renderer.Render(rpt))
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

func (m Map) render(rpt report.Report) (RenderableNodes, map[string]report.IDList) {
	input := m.Renderer.Render(rpt)
	output := RenderableNodes{}
	mapped := map[string]report.IDList{}      // input node ID -> output node IDs
	adjacencies := map[string]report.IDList{} // output node ID -> input node Adjacencies

	for _, inRenderable := range input {
		outRenderables := m.MapFunc(inRenderable)
		for _, outRenderable := range outRenderables {
			existing, ok := output[outRenderable.ID]
			if ok {
				outRenderable.Merge(existing)
			}

			output[outRenderable.ID] = outRenderable
			mapped[inRenderable.ID] = mapped[inRenderable.ID].Add(outRenderable.ID)
			adjacencies[outRenderable.ID] = adjacencies[outRenderable.ID].Merge(inRenderable.Adjacency)
		}
	}

	// Rewrite Adjacency for new node IDs.
	// NB we don't do pseudo nodes here; we assume the input graph
	// is properly-connected, and if the map func dropped a node,
	// we drop links to it.
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

// LeafMap is a Renderer which produces a set of RenderableNodes from a report.Topology
// by using a map function and topology selector.
type LeafMap struct {
	Selector report.TopologySelector
	Mapper   LeafMapFunc
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
		source2mapped = map[string]report.IDList{} // input node ID -> output node IDs
		adjacencies   = map[string]report.IDList{} // input node ID -> input node Adjacencies
		localNetworks = LocalNetworks(rpt)
	)

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	for nodeID, metadata := range t.NodeMetadatas {
		for _, mapped := range m.Mapper(metadata, localNetworks) {
			// mapped.ID needs not be unique over all addressIDs. If not, we merge with
			// the existing data, on the assumption that the MapFunc returns the same
			// data.
			existing, ok := nodes[mapped.ID]
			if ok {
				mapped.Merge(existing)
			}

			origins := mapped.Origins
			origins = origins.Add(nodeID)
			if hostNodeID, ok := metadata.Metadata[report.HostNodeID]; ok {
				origins = origins.Add(hostNodeID)
			}
			mapped.Origins = origins
			nodes[mapped.ID] = mapped
			source2mapped[nodeID] = source2mapped[nodeID].Add(mapped.ID)
			adjacencies[nodeID] = metadata.Adjacency
		}
	}

	// We propagate edge metadata to nodes on both ends of the edges.
	// TODO we should 'reverse' one end of the edge meta data - ingress -> egress etc.
	for srcNodeID, nmd := range t.NodeMetadatas {
		for _, srcRenderableID := range source2mapped[srcNodeID] {
			srcRenderableNode := nodes[srcRenderableID]

			for dstNodeID, emd := range nmd.Edges {
				for _, dstRenderableID := range source2mapped[dstNodeID] {
					dstRenderableNode := nodes[dstRenderableID]

					srcRenderableNode.EdgeMetadata = srcRenderableNode.EdgeMetadata.Merge(emd)
					dstRenderableNode.EdgeMetadata = dstRenderableNode.EdgeMetadata.Merge(emd)

					nodes[dstRenderableID] = dstRenderableNode
				}
			}

			nodes[srcRenderableID] = srcRenderableNode
		}
	}

	// Walk the graph and make connections.
	for srcNodeID, dstNodeIDs := range adjacencies {
		for _, srcRenderableID := range source2mapped[srcNodeID] {
			srcRenderableNode := nodes[srcRenderableID]

			for _, dstNodeID := range dstNodeIDs {
				for _, dstRenderableID := range source2mapped[dstNodeID] {
					srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstRenderableID)
				}
			}

			nodes[srcRenderableID] = srcRenderableNode
		}
	}

	return nodes
}

func ids(nodes RenderableNodes) report.IDList {
	result := report.MakeIDList()
	for id := range nodes {
		result = result.Add(id)
	}
	return result
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (m LeafMap) EdgeMetadata(rpt report.Report, srcRenderableID, dstRenderableID string) report.EdgeMetadata {
	var (
		t             = m.Selector(rpt)
		localNetworks = LocalNetworks(rpt)
		metadata      = report.EdgeMetadata{}
	)
	for src, nmd := range t.NodeMetadatas {
		for dst, edgeMeta := range nmd.Edges {
			srcs, dsts := report.MakeIDList(src), report.MakeIDList(dst)
			if src != report.TheInternet {
				mapped := m.Mapper(t.NodeMetadatas[src], localNetworks)
				srcs = ids(mapped)
			}
			if dst != report.TheInternet {
				mapped := m.Mapper(t.NodeMetadatas[dst], localNetworks)
				dsts = ids(mapped)
			}
			if srcs.Contains(srcRenderableID) && dsts.Contains(dstRenderableID) {
				metadata = metadata.Flatten(edgeMeta)
			}
		}
	}
	return metadata
}

// CustomRenderer allow for mapping functions that recived the entire topology
// in one call - useful for functions that need to consider the entire graph
type CustomRenderer struct {
	RenderFunc func(RenderableNodes) RenderableNodes
	Renderer
}

// Render implements Renderer
func (c CustomRenderer) Render(rpt report.Report) RenderableNodes {
	return c.RenderFunc(c.Renderer.Render(rpt))
}

// IsConnected is the key added to NodeMetadata by ColorConnected
// to indicate a node has an edge pointing to it or from it
const IsConnected = "is_connected"

// OnlyConnected filters out unconnected RenderedNodes
func OnlyConnected(input RenderableNodes) RenderableNodes {
	output := RenderableNodes{}
	for id, node := range ColorConnected(input) {
		if _, ok := node.NodeMetadata.Metadata[IsConnected]; ok {
			output[id] = node
		}
	}
	return output
}

// FilterUnconnected produces a renderer that filters unconnected nodes
// from the given renderer
func FilterUnconnected(r Renderer) Renderer {
	return CustomRenderer{
		RenderFunc: OnlyConnected,
		Renderer:   r,
	}
}

// ColorConnected colors nodes with the IsConnected key if
// they have edges to or from them.
func ColorConnected(input RenderableNodes) RenderableNodes {
	connected := map[string]struct{}{}
	void := struct{}{}

	for id, node := range input {
		if len(node.Adjacency) == 0 {
			continue
		}

		connected[id] = void
		for _, id := range node.Adjacency {
			connected[id] = void
		}
	}

	for id := range connected {
		node := input[id]
		node.NodeMetadata.Metadata[IsConnected] = "true"
		input[id] = node
	}
	return input
}

// Filter removes nodes from a view based on a predicate.
type Filter struct {
	Renderer
	f func(RenderableNode) bool
}

// Render implements Renderer
func (f Filter) Render(rpt report.Report) RenderableNodes {
	output := RenderableNodes{}
	for id, node := range f.Renderer.Render(rpt) {
		if f.f(node) {
			output[id] = node
		}
	}
	return output
}
