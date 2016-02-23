package detailed

import (
	"sort"
	"strconv"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

const (
	portKey    = "port"
	portLabel  = "Port"
	countKey   = "count"
	countLabel = "Count"
	number     = "number"
)

// Exported for testing
var (
	NormalColumns = []Column{
		{ID: portKey, Label: portLabel},
		{ID: countKey, Label: countLabel, DefaultSort: true},
	}
	InternetColumns = []Column{
		{ID: "foo", Label: "Remote"},
		{ID: portKey, Label: portLabel},
		{ID: countKey, Label: countLabel, DefaultSort: true},
	}
)

func endpointChildrenOf(n render.RenderableNode) []render.RenderableNode {
	result := []render.RenderableNode{}
	n.Children.ForEach(func(child render.RenderableNode) {
		if _, _, _, ok := render.ParseEndpointID(child.ID); ok {
			result = append(result, child)
		}
	})
	return result
}

func endpointChildIDsOf(n render.RenderableNode) report.IDList {
	result := report.MakeIDList()
	n.Children.ForEach(func(child render.RenderableNode) {
		if _, _, _, ok := render.ParseEndpointID(child.ID); ok {
			result = append(result, child.ID)
		}
	})
	return result
}

type connectionsRow struct {
	remoteNode, localNode *render.RenderableNode
	remoteAddr, localAddr string
	port                  string // always the server-side port
}

func buildConnectionRows(in map[connectionsRow]int, includeLocal bool) []NodeSummary {
	nodes := []NodeSummary{}
	for row, count := range in {
		id, label, linkable := row.remoteNode.ID, row.remoteNode.LabelMajor, true
		if row.remoteAddr != "" {
			id, label, linkable = row.remoteAddr+":"+row.port, row.remoteAddr, false
		}
		metadata := []MetadataRow{}
		if includeLocal {
			metadata = append(metadata,
				MetadataRow{
					ID:       "foo",
					Value:    row.localAddr,
					Datatype: number,
				})
		}
		metadata = append(metadata,
			MetadataRow{
				ID:       portKey,
				Value:    row.port,
				Datatype: number,
			},
			MetadataRow{
				ID:       countKey,
				Value:    strconv.Itoa(count),
				Datatype: number,
			},
		)
		nodes = append(nodes, NodeSummary{
			ID:       id,
			Label:    label,
			Linkable: linkable,
			Metadata: metadata,
		})
	}
	sort.Sort(nodeSummariesByID(nodes))
	return nodes
}

func isInternetNode(n render.RenderableNode) bool {
	return n.ID == render.IncomingInternetID || n.ID == render.OutgoingInternetID
}

func makeIncomingConnectionsTable(topologyID string, n render.RenderableNode, ns render.RenderableNodes) NodeSummaryGroup {
	localEndpointIDs := endpointChildIDsOf(n)

	// For each node which has an edge TO me
	counts := map[connectionsRow]int{}
	for _, node := range ns {
		if !node.Adjacency.Contains(n.ID) {
			continue
		}
		remoteNode := node.Copy()

		// Work out what port they are talking to, and count the number of
		// connections to that port.
		// This is complicated as for internet nodes we break out individual
		// address, both when the internet node is remote (an incoming
		// connection from the internet) and 'local' (ie you are loading
		// details on the internet node)
		for _, child := range endpointChildrenOf(node) {
			for _, localEndpointID := range child.Adjacency.Intersection(localEndpointIDs) {
				_, localAddr, port, ok := render.ParseEndpointID(localEndpointID)
				if !ok {
					continue
				}
				key := connectionsRow{
					localNode:  &n,
					remoteNode: &remoteNode,
					port:       port,
				}
				if isInternetNode(n) {
					key.localAddr = localAddr
				}
				counts[key] = counts[key] + 1
			}
		}
	}

	columnHeaders := NormalColumns
	if isInternetNode(n) {
		columnHeaders = InternetColumns
	}
	return NodeSummaryGroup{
		ID:         "incoming-connections",
		TopologyID: topologyID,
		Label:      "Inbound",
		Columns:    columnHeaders,
		Nodes:      buildConnectionRows(counts, isInternetNode(n)),
	}
}

func makeOutgoingConnectionsTable(topologyID string, n render.RenderableNode, ns render.RenderableNodes) NodeSummaryGroup {
	localEndpoints := endpointChildrenOf(n)

	// For each node which has an edge FROM me
	counts := map[connectionsRow]int{}
	for _, node := range ns {
		if !n.Adjacency.Contains(node.ID) {
			continue
		}
		remoteNode := node.Copy()
		remoteEndpointIDs := endpointChildIDsOf(remoteNode)

		for _, localEndpoint := range localEndpoints {
			_, localAddr, _, ok := render.ParseEndpointID(localEndpoint.ID)
			if !ok {
				continue
			}

			for _, remoteEndpointID := range localEndpoint.Adjacency.Intersection(remoteEndpointIDs) {
				_, _, port, ok := render.ParseEndpointID(remoteEndpointID)
				if !ok {
					continue
				}
				key := connectionsRow{
					localNode:  &n,
					remoteNode: &remoteNode,
					port:       port,
				}
				if isInternetNode(n) {
					key.localAddr = localAddr
				}
				counts[key] = counts[key] + 1
			}
		}
	}

	columnHeaders := NormalColumns
	if isInternetNode(n) {
		columnHeaders = InternetColumns
	}
	return NodeSummaryGroup{
		ID:         "outgoing-connections",
		TopologyID: topologyID,
		Label:      "Outbound",
		Columns:    columnHeaders,
		Nodes:      buildConnectionRows(counts, isInternetNode(n)),
	}
}
