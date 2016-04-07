package detailed

import (
	"fmt"
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

// ConnectionsTable is the table of connection to/form a node
type ConnectionsTable struct {
	ID         string           `json:"id"`
	TopologyID string           `json:"topology_id"`
	Label      string           `json:"label"`
	Columns    []Column         `json:"columns"`
	Rows       []ConnectionsRow `json:"rows"`
}

// ConnectionsRow is a row in the connections table.
type ConnectionsRow struct {
	ID       string        `json:"id"`
	NodeID   string        `json:"node_id"`
	Label    string        `json:"label"`
	Linkable bool          `json:"linkable"`
	Metadata []MetadataRow `json:"metadata,omitempty"`
}

type connectionsRowsByID []ConnectionsRow

func (s connectionsRowsByID) Len() int           { return len(s) }
func (s connectionsRowsByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s connectionsRowsByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

// Intermediate type used as a key to dedupe rows
type connectionsRow struct {
	remoteNode, localNode *report.Node
	remoteAddr, localAddr string
	port                  string // always the server-side port
}

func (row connectionsRow) ID() string {
	return fmt.Sprintf("%s:%s-%s:%s-%s", row.remoteNode.ID, row.remoteAddr, row.localNode.ID, row.localAddr, row.port)
}

func incomingConnectionsTable(topologyID string, n report.Node, ns report.Nodes) ConnectionsTable {
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
				_, localAddr, port, ok := report.ParseEndpointNodeID(localEndpointID)
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
	return ConnectionsTable{
		ID:         "incoming-connections",
		TopologyID: topologyID,
		Label:      "Inbound",
		Columns:    columnHeaders,
		Rows:       connectionRows(counts, isInternetNode(n)),
	}
}

func outgoingConnectionsTable(topologyID string, n report.Node, ns report.Nodes) ConnectionsTable {
	localEndpoints := endpointChildrenOf(n)

	// For each node which has an edge FROM me
	counts := map[connectionsRow]int{}
	for _, id := range n.Adjacency {
		node, ok := ns[id]
		if !ok {
			continue
		}
		remoteNode := node.Copy()
		remoteEndpointIDs := endpointChildIDsOf(remoteNode)

		for _, localEndpoint := range localEndpoints {
			_, localAddr, _, ok := report.ParseEndpointNodeID(localEndpoint.ID)
			if !ok {
				continue
			}

			for _, remoteEndpointID := range localEndpoint.Adjacency.Intersection(remoteEndpointIDs) {
				_, _, port, ok := report.ParseEndpointNodeID(remoteEndpointID)
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
	return ConnectionsTable{
		ID:         "outgoing-connections",
		TopologyID: topologyID,
		Label:      "Outbound",
		Columns:    columnHeaders,
		Rows:       connectionRows(counts, isInternetNode(n)),
	}
}

func endpointChildrenOf(n report.Node) []report.Node {
	result := []report.Node{}
	n.Children.ForEach(func(child report.Node) {
		if child.Topology == report.Endpoint {
			result = append(result, child)
		}
	})
	return result
}

func endpointChildIDsOf(n report.Node) report.IDList {
	result := report.MakeIDList()
	n.Children.ForEach(func(child report.Node) {
		if child.Topology == report.Endpoint {
			result = result.Add(child.ID)
		}
	})
	return result
}

func isInternetNode(n report.Node) bool {
	return n.ID == render.IncomingInternetID || n.ID == render.OutgoingInternetID
}

func connectionRows(in map[connectionsRow]int, includeLocal bool) []ConnectionsRow {
	output := []ConnectionsRow{}
	for row, count := range in {
		// Use MakeNodeSummary to render the id and label of this node
		// TODO(paulbellamy): Would be cleaner if we hade just a
		// MakeNodeID(*row.remoteode). As we don't need the whole summary.
		summary, ok := MakeNodeSummary(*row.remoteNode)
		connectionsRow := ConnectionsRow{
			ID:       row.ID(),
			NodeID:   summary.ID,
			Label:    summary.Label,
			Linkable: true,
		}
		if !ok && row.remoteAddr != "" {
			connectionsRow.Label = row.remoteAddr
			connectionsRow.Linkable = false
		}
		if includeLocal {
			connectionsRow.Metadata = append(connectionsRow.Metadata,
				MetadataRow{
					ID:       "foo",
					Value:    row.localAddr,
					Datatype: number,
				})
		}
		connectionsRow.Metadata = append(connectionsRow.Metadata,
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
		output = append(output, connectionsRow)
	}
	sort.Sort(connectionsRowsByID(output))
	return output
}
