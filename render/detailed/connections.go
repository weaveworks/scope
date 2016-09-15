package detailed

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/weaveworks/scope/probe/endpoint"
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

// ConnectionsSummary is the table of connection to/form a node
type ConnectionsSummary struct {
	ID          string       `json:"id"`
	TopologyID  string       `json:"topologyId"`
	Label       string       `json:"label"`
	Columns     []Column     `json:"columns"`
	Connections []Connection `json:"connections"`
}

// Connection is a row in the connections table.
type Connection struct {
	ID       string               `json:"id"`     // ID of this element in the UI.  Must be unique for a given ConnectionsSummary.
	NodeID   string               `json:"nodeId"` // ID of a node in the topology. Optional, must be set if linkable is true.
	Label    string               `json:"label"`
	Linkable bool                 `json:"linkable"`
	Metadata []report.MetadataRow `json:"metadata,omitempty"`
}

type connectionsByID []Connection

func (s connectionsByID) Len() int           { return len(s) }
func (s connectionsByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s connectionsByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

// Intermediate type used as a key to dedupe rows
type connection struct {
	remoteNodeID, localNodeID string
	port                      string // always the server-side port
}

func newConnection(n report.Node, node report.Node, port string, endpointID string) connection {
	c := connection{
		localNodeID:  n.ID,
		remoteNodeID: node.ID,
		port:         port,
	}
	// For internet nodes we break out individual addresses, both when
	// the internet node is remote (an incoming connection from the
	// internet) and 'local' (ie you are loading details on the
	// internet node). Hence we use the *endpoint* ID here since that
	// gives us the address and reverse DNS information.
	if isInternetNode(n) {
		c.localNodeID = endpointID
	}
	return c
}

func (row connection) ID() string {
	return fmt.Sprintf("%s-%s-%s", row.remoteNodeID, row.localNodeID, row.port)
}

type connectionCounters struct {
	counted map[string]struct{}
	counts  map[connection]int
}

func newConnectionCounters() *connectionCounters {
	return &connectionCounters{counted: map[string]struct{}{}, counts: map[connection]int{}}
}

func (c *connectionCounters) add(sourceEndpoint report.Node, n report.Node, node report.Node, port string, endpointID string) {
	// We identify connections by their source endpoint, pre-NAT, to
	// ensure we only count them once.
	connectionID := sourceEndpoint.ID
	if copySourceEndpointID, _, ok := sourceEndpoint.Latest.LookupEntry("copy_of"); ok {
		connectionID = copySourceEndpointID
	}
	if _, ok := c.counted[connectionID]; ok {
		return
	}
	c.counted[connectionID] = struct{}{}
	key := newConnection(n, node, port, endpointID)
	c.counts[key]++
}

func (c *connectionCounters) rows(r report.Report, ns report.Nodes, includeLocal bool) []Connection {
	output := []Connection{}
	for row, count := range c.counts {
		// Use MakeNodeSummary to render the id and label of this node
		// TODO(paulbellamy): Would be cleaner if we hade just a
		// MakeNodeID(ns[row.remoteNodeID]). As we don't need the whole summary.
		summary, _ := MakeNodeSummary(r, ns[row.remoteNodeID])
		connection := Connection{
			ID:       row.ID(),
			NodeID:   summary.ID,
			Label:    summary.Label,
			Linkable: true,
		}
		if includeLocal {
			_, label, _, _ := report.ParseEndpointNodeID(row.localNodeID)
			// Does localNode (which, in this case, is an endpoint)
			// have a DNS record in it?
			if set, ok := r.Endpoint.Nodes[row.localNodeID].Sets.Lookup(endpoint.ReverseDNSNames); ok && len(set) > 0 {
				label = fmt.Sprintf("%s (%s)", set[0], label)
			}
			connection.Metadata = append(connection.Metadata,
				report.MetadataRow{
					ID:       "foo",
					Value:    label,
					Datatype: number,
				})
		}
		connection.Metadata = append(connection.Metadata,
			report.MetadataRow{
				ID:       portKey,
				Value:    row.port,
				Datatype: number,
			},
			report.MetadataRow{
				ID:       countKey,
				Value:    strconv.Itoa(count),
				Datatype: number,
			},
		)
		output = append(output, connection)
	}
	sort.Sort(connectionsByID(output))
	return output
}

func incomingConnectionsSummary(topologyID string, r report.Report, n report.Node, ns report.Nodes) ConnectionsSummary {
	localEndpointIDs, localEndpointIDCopies := endpointChildIDsAndCopyMapOf(n)
	counts := newConnectionCounters()

	// For each node which has an edge TO me
	for _, node := range ns {
		if !node.Adjacency.Contains(n.ID) {
			continue
		}
		// Work out what port they are talking to, and count the number of
		// connections to that port.
		for _, remoteEndpoint := range endpointChildrenOf(node) {
			for _, localEndpointID := range remoteEndpoint.Adjacency.Intersection(localEndpointIDs) {
				localEndpointID = canonicalEndpointID(localEndpointIDCopies, localEndpointID)
				_, _, port, ok := report.ParseEndpointNodeID(localEndpointID)
				if !ok {
					continue
				}
				counts.add(remoteEndpoint, n, node, port, localEndpointID)
			}
		}
	}

	columnHeaders := NormalColumns
	if isInternetNode(n) {
		columnHeaders = InternetColumns
	}
	return ConnectionsSummary{
		ID:          "incoming-connections",
		TopologyID:  topologyID,
		Label:       "Inbound",
		Columns:     columnHeaders,
		Connections: counts.rows(r, ns, isInternetNode(n)),
	}
}

func outgoingConnectionsSummary(topologyID string, r report.Report, n report.Node, ns report.Nodes) ConnectionsSummary {
	localEndpoints := endpointChildrenOf(n)
	counts := newConnectionCounters()

	// For each node which has an edge FROM me
	for _, id := range n.Adjacency {
		node, ok := ns[id]
		if !ok {
			continue
		}

		remoteEndpointIDs, remoteEndpointIDCopies := endpointChildIDsAndCopyMapOf(node)

		for _, localEndpoint := range localEndpoints {
			for _, remoteEndpointID := range localEndpoint.Adjacency.Intersection(remoteEndpointIDs) {
				remoteEndpointID = canonicalEndpointID(remoteEndpointIDCopies, remoteEndpointID)
				_, _, port, ok := report.ParseEndpointNodeID(remoteEndpointID)
				if !ok {
					continue
				}
				counts.add(localEndpoint, n, node, port, localEndpoint.ID)
			}
		}
	}

	columnHeaders := NormalColumns
	if isInternetNode(n) {
		columnHeaders = InternetColumns
	}
	return ConnectionsSummary{
		ID:          "outgoing-connections",
		TopologyID:  topologyID,
		Label:       "Outbound",
		Columns:     columnHeaders,
		Connections: counts.rows(r, ns, isInternetNode(n)),
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

func endpointChildIDsAndCopyMapOf(n report.Node) (report.IDList, map[string]string) {
	ids := report.MakeIDList()
	copies := map[string]string{}
	n.Children.ForEach(func(child report.Node) {
		if child.Topology == report.Endpoint {
			ids = ids.Add(child.ID)
			if copyID, _, ok := child.Latest.LookupEntry("copy_of"); ok {
				copies[child.ID] = copyID
			}
		}
	})
	return ids, copies
}

// canonicalEndpointID returns the original endpoint ID of which id is
// a "copy_of" (due to NATing), or, if the id is not a copy, the id
// itself.
//
// This is used for determining a unique destination endpoint ID for a
// connection, removing any arbitrariness in the destination port we
// are associating with the connection when it is encountered multiple
// times in the topology (with different destination endpoints, due to
// DNATing).
func canonicalEndpointID(copies map[string]string, id string) string {
	if original, ok := copies[id]; ok {
		return original
	}
	return id
}

func isInternetNode(n report.Node) bool {
	return n.ID == render.IncomingInternetID || n.ID == render.OutgoingInternetID
}
