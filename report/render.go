package report

import "fmt"

// Render transforms a report into a set of RenderableNodes, which the UI will
// render collectively as a graph. RenderBy takes a MapFunc, which defines how
// to group and label nodes. Nodes with the same mapped IDs will be merged.
func Render(r Report, ts TopologySelector, mapper MapFunc, pseudo PseudoFunc) map[string]RenderableNode {
	var (
		nodes = map[string]RenderableNode{}
		t     = ts(r)
	)

	// Build a set of RenderableNodes for all node IDs, and an ID lookup map.
	// Multiple IDs can map to the same RenderableNodes.
	source2mapped := map[string]string{}
	for nodeID := range t.NodeMetadatas {
		mapped, show := mapper(r, ts, nodeID)
		if !show {
			continue
		}

		// mapped.ID needs not be unique over all IDs. We just overwrite the
		// existing data, on the assumption that the MapFunc returns the same
		// data.
		nodes[mapped.ID] = RenderableNode{
			ID:         mapped.ID,
			LabelMajor: mapped.Major,
			LabelMinor: mapped.Minor,
			Rank:       mapped.Rank,
			Pseudo:     false,
			Adjacency:  IDList{},            // later
			Origins:    IDList{},            // later
			Metadata:   AggregateMetadata{}, // later
		}
		source2mapped[nodeID] = mapped.ID
	}

	// Walk the graph and make connections.
	for adjacencyID, dstNodeIDs := range t.Adjacency {
		srcHostID, srcNodeID, ok := ParseAdjacencyID(adjacencyID)
		if !ok {
			panic(fmt.Sprintf("badly formed Topology: invalid adjacencyID %q", adjacencyID))
		}

		srcMappedNodeID, ok := source2mapped[srcNodeID]
		if !ok {
			if pseudo == nil {
				continue // oh well
			}
			pseudoNode := pseudo(srcNodeID)
			srcMappedNodeID = pseudoNode.ID
			source2mapped[srcNodeID] = srcMappedNodeID
			nodes[srcMappedNodeID] = pseudo2renderable(pseudoNode)
		}

		srcRenderableNode, ok := nodes[srcMappedNodeID]
		if !ok {
			panic(fmt.Sprintf("badly formed mapping: %q (via %q) has no renderable node", srcMappedNodeID, srcNodeID))
		}

		for _, dstNodeID := range dstNodeIDs {
			dstMappedNodeID, ok := source2mapped[dstNodeID]
			if !ok {
				if pseudo == nil {
					continue // oh well
				}
				pseudoNode := pseudo(dstNodeID)
				dstMappedNodeID = pseudoNode.ID
				source2mapped[dstNodeID] = dstMappedNodeID
				nodes[dstMappedNodeID] = pseudo2renderable(pseudoNode)
			}

			srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstMappedNodeID)
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(MakeHostNodeID(srcHostID))
			srcRenderableNode.Origins = srcRenderableNode.Origins.Add(srcNodeID)
			edgeID := MakeEdgeID(srcNodeID, dstNodeID)
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				srcRenderableNode.Metadata = srcRenderableNode.Metadata.Merge(md.Export())
			}
		}

		nodes[srcMappedNodeID] = srcRenderableNode
	}

	return nodes
}

// RenderableNode is the data type that's yielded to the JavaScript layer as
// an element of a topology. It should contain information that's relevant
// to rendering a node when there are many nodes visible at once.
type RenderableNode struct {
	ID         string            `json:"id"`                    //
	LabelMajor string            `json:"label_major"`           // e.g. "process", human-readable
	LabelMinor string            `json:"label_minor,omitempty"` // e.g. "hostname", human-readable, optional
	Rank       string            `json:"rank"`                  // to help the layout engine
	Pseudo     bool              `json:"pseudo,omitempty"`      // sort-of a placeholder node, for rendering purposes
	Adjacency  IDList            `json:"adjacency,omitempty"`   // Same-topology node IDs
	Origins    IDList            `json:"origins,omitempty"`     // Foreign-key-scoped, core node IDs that contributed to this RenderableNode
	Metadata   AggregateMetadata `json:"metadata"`              // Numeric sums
}

func pseudo2renderable(mappedNode MappedNode) RenderableNode {
	return RenderableNode{
		ID:         mappedNode.ID,
		LabelMajor: mappedNode.Major,
		LabelMinor: mappedNode.Minor,
		Rank:       mappedNode.Rank,
		Pseudo:     true,
		Adjacency:  IDList{},            // fill in later
		Origins:    IDList{},            // fill in later
		Metadata:   AggregateMetadata{}, // fill in later
	}
}
