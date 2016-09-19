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
	portKey     = "port"
	portLabel   = "Port"
	countKey    = "count"
	countLabel  = "Count"
	remoteKey   = "remote"
	remoteLabel = "Remote"
	number      = "number"
)

// Exported for testing
var (
	NormalColumns = []Column{
		{ID: portKey, Label: portLabel},
		{ID: countKey, Label: countLabel, DefaultSort: true},
	}
	InternetColumns = []Column{
		{ID: remoteKey, Label: remoteLabel},
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
	remoteNodeID          string
	remoteAddr, localAddr string // for internet nodes only
	port                  string // destination port
}

type connectionCounters struct {
	counted map[string]struct{}
	counts  map[connection]int
}

func newConnectionCounters() *connectionCounters {
	return &connectionCounters{counted: map[string]struct{}{}, counts: map[connection]int{}}
}

func (c *connectionCounters) add(outgoing bool, localNode, remoteNode, localEndpoint, remoteEndpoint report.Node) {
	// We identify connections by their source endpoint, pre-NAT, to
	// ensure we only count them once.
	srcEndpoint, dstEndpoint := remoteEndpoint, localEndpoint
	if outgoing {
		srcEndpoint, dstEndpoint = localEndpoint, remoteEndpoint
	}
	connectionID := srcEndpoint.ID
	if copySrcEndpointID, _, ok := srcEndpoint.Latest.LookupEntry("copy_of"); ok {
		connectionID = copySrcEndpointID
	}
	if _, ok := c.counted[connectionID]; ok {
		return
	}

	conn := connection{remoteNodeID: remoteNode.ID}
	var ok bool
	if _, _, conn.port, ok = report.ParseEndpointNodeID(dstEndpoint.ID); !ok {
		return
	}
	// For internet nodes we break out individual addresses
	if conn.remoteAddr, ok = internetAddr(remoteNode, remoteEndpoint); !ok {
		return
	}
	if conn.localAddr, ok = internetAddr(localNode, localEndpoint); !ok {
		return
	}

	c.counted[connectionID] = struct{}{}
	c.counts[conn]++
}

func internetAddr(node report.Node, ep report.Node) (string, bool) {
	if !isInternetNode(node) {
		return "", true
	}
	_, addr, _, ok := report.ParseEndpointNodeID(ep.ID)
	if !ok {
		return "", false
	}
	if set, ok := ep.Sets.Lookup(endpoint.ReverseDNSNames); ok && len(set) > 0 {
		// TODO We show just one of the names, selected rather
		// abitrarily. We don't have space to show all (except in the
		// tooltip perhaps), but should think of better strategies for
		// choosing the name to display.
		addr = fmt.Sprintf("%s (%s)", set[0], addr)
	}
	return addr, true
}

func (c *connectionCounters) rows(r report.Report, ns report.Nodes, includeLocal bool) []Connection {
	output := []Connection{}
	for row, count := range c.counts {
		// Use MakeNodeSummary to render the id and label of this node
		// TODO(paulbellamy): Would be cleaner if we hade just a
		// MakeNodeID(ns[row.remoteNodeID]). As we don't need the whole summary.
		summary, _ := MakeNodeSummary(r, ns[row.remoteNodeID])
		connection := Connection{
			ID:       fmt.Sprintf("%s-%s-%s-%s", row.remoteNodeID, row.remoteAddr, row.localAddr, row.port),
			NodeID:   summary.ID,
			Label:    summary.Label,
			Linkable: true,
		}
		if row.remoteAddr != "" {
			connection.Label = row.remoteAddr
		}
		if includeLocal {
			connection.Metadata = append(connection.Metadata,
				report.MetadataRow{
					ID:       remoteKey,
					Value:    row.localAddr,
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
		for _, remoteEndpoint := range endpointChildrenOf(node) {
			for _, localEndpointID := range remoteEndpoint.Adjacency.Intersection(localEndpointIDs) {
				localEndpointID = canonicalEndpointID(localEndpointIDCopies, localEndpointID)
				counts.add(false, n, node, r.Endpoint.Nodes[localEndpointID], remoteEndpoint)
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
				counts.add(true, n, node, localEndpoint, r.Endpoint.Nodes[remoteEndpointID])
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
