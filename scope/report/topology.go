package report

import (
	"net"
	"reflect"
	"strings"
)

const (
	// ScopeDelim separates the scope portion of an address from the address
	// string itself.
	ScopeDelim = ";"

	// IDDelim separates fields in a node ID.
	IDDelim = "|"

	localUnknown = "localUnknown"
)

// Topology describes a specific view of a network. It consists of nodes and
// edges, represented by Adjacency, and metadata about those nodes and edges,
// represented by EdgeMetadatas and NodeMetadatas respectively.
type Topology struct {
	Adjacency
	EdgeMetadatas
	NodeMetadatas
}

// Adjacency is an adjacency-list encoding of the topology. Keys are node IDs,
// as produced by the relevant MappingFunc for the topology.
type Adjacency map[string]IDList

// EdgeMetadatas collect metadata about each edge in a topology. Keys are a
// concatenation of node IDs.
type EdgeMetadatas map[string]EdgeMetadata

// NodeMetadatas collect metadata about each node in a topology. Keys are node
// IDs.
type NodeMetadatas map[string]NodeMetadata

// EdgeMetadata describes a superset of the metadata that probes can
// conceivably (and usefully) collect about an edge between two nodes in any
// topology.
type EdgeMetadata struct {
	WithBytes    bool
	BytesIngress uint // dst -> src
	BytesEgress  uint // src -> dst

	WithConnCountTCP bool
	MaxConnCountTCP  uint
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology. Right now it's a weakly-typed map,
// which should probably change (see comment on type MapFunc).
type NodeMetadata map[string]string

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// RenderBy transforms a given Topology into a set of RenderableNodes, which
// the UI will render collectively as a graph. It takes a a MapFunc, which
// defines how to group and label nodes. If grouped is true, nodes that belong
// to the same "class" will be merged into a single RenderableNode.
func (t Topology) RenderBy(f MapFunc, grouped bool) map[string]RenderableNode {
	nodes := map[string]RenderableNode{}

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	address2mapped := map[string]string{}
	for addressID, meta := range t.NodeMetadatas {
		mapped, ok := f(addressID, meta, grouped)
		if !ok {
			continue
		}

		// ID needs not be unique.
		nodes[mapped.ID] = RenderableNode{
			ID:         mapped.ID,
			LabelMajor: mapped.Major,
			LabelMinor: mapped.Minor,
			Rank:       mapped.Rank,
			Pseudo:     false,
			Metadata:   AggregateMetadata{}, // can only fill in later
		}

		address2mapped[addressID] = mapped.ID
	}

	// Walk the graph and make connections.
	for local, remotes := range t.Adjacency {
		var (
			fields       = strings.SplitN(local, IDDelim, 2) // "<host>|<address>"
			originID     = fields[0]
			localAddress = fields[1]
			localID      = address2mapped[localAddress] // must exist
			localNode    = nodes[localID]               // must exist
		)

		for _, remoteAddress := range remotes {
			remoteID, ok := address2mapped[remoteAddress]
			if !ok {
				// We don't have a node for this target address. So we'll make
				// a pseudonode for it, instead.
				var maj, min string
				if remoteAddress == TheInternet {
					remoteID = remoteAddress
					maj, min = formatLabel(remoteAddress)
				} else if grouped {
					remoteID = localUnknown
					maj, min = "", ""
				} else {
					remoteID = "pseudo:" + remoteAddress
					maj, min = formatLabel(remoteAddress)
				}
				nodes[remoteID] = RenderableNode{
					ID:         remoteID,
					LabelMajor: maj,
					LabelMinor: min,
					Pseudo:     true,
					Metadata:   AggregateMetadata{}, // populated below
				}
				address2mapped[remoteAddress] = remoteID
			}
			localNode.Origin = localNode.Origin.Add(originID)
			localNode.Adjacency = localNode.Adjacency.Add(remoteID)

			edgeID := localAddress + IDDelim + remoteAddress
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				localNode.Metadata.Merge(md.Render())
			}
		}

		nodes[localID] = localNode
	}

	return nodes
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// localMappedID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate addressIDs to
// mappedIDs.
func (t Topology) EdgeMetadata(f MapFunc, grouped bool, localMappedID, remoteMappedID string) EdgeMetadata {
	metadata := EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		edgeParts := strings.SplitN(edgeID, IDDelim, 2)
		localID := edgeParts[0]
		if localID != TheInternet {
			mapped, _ := f(localID, t.NodeMetadatas[localID], grouped)
			localID = mapped.ID
		}
		remoteID := edgeParts[1]
		if remoteID != TheInternet {
			mapped, _ := f(remoteID, t.NodeMetadatas[remoteID], grouped)
			remoteID = mapped.ID
		}
		if localID == localMappedID && remoteID == remoteMappedID {
			metadata.Flatten(edgeMeta)
		}
	}
	return metadata
}

// formatLabel is an opportunistic helper to format any addressID into
// something we can show on screen.
func formatLabel(s string) (string, string) {
	if s == TheInternet {
		return "the Internet", ""
	}

	// Format is either "scope;ip;port", "scope;ip", or some process id.
	parts := strings.SplitN(s, ScopeDelim, 3)
	if len(parts) < 2 {
		return s, ""
	}

	if len(parts) == 2 {
		return parts[1], ""
	}

	return net.JoinHostPort(parts[1], parts[2]), ""
}

// Diff is returned by TopoDiff. It represents the changes between two
// RenderableNode maps.
type Diff struct {
	Add    []RenderableNode `json:"add"`
	Update []RenderableNode `json:"update"`
	Remove []string         `json:"remove"`
}

// TopoDiff gives you the diff to get from A to B.
func TopoDiff(a, b map[string]RenderableNode) Diff {
	diff := Diff{}

	notSeen := map[string]struct{}{}
	for k := range a {
		notSeen[k] = struct{}{}
	}

	for k, node := range b {
		if _, ok := a[k]; !ok {
			diff.Add = append(diff.Add, node)
		} else {
			if !reflect.DeepEqual(node, a[k]) {
				diff.Update = append(diff.Update, node)
			}
		}
		delete(notSeen, k)
	}

	// leftover keys
	for k := range notSeen {
		diff.Remove = append(diff.Remove, k)
	}

	return diff
}

// ByID is a sort interface for a RenderableNode slice.
type ByID []RenderableNode

func (r ByID) Len() int           { return len(r) }
func (r ByID) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByID) Less(i, j int) bool { return r[i].ID < r[j].ID }
