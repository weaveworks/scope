package report

import (
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
	WithBytes    bool `json:"with_bytes,omitempty"`
	BytesIngress uint `json:"bytes_ingress,omitempty"` // dst -> src
	BytesEgress  uint `json:"bytes_egress,omitempty"`  // src -> dst

	WithConnCountTCP bool `json:"with_conn_count_tcp,omitempty"`
	MaxConnCountTCP  uint `json:"max_conn_count_tcp,omitempty"`
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology. Right now it's a weakly-typed map,
// which should probably change (see comment on type MapFunc).
type NodeMetadata map[string]string

// Copy returns a value copy, useful for tests.
func (nm NodeMetadata) Copy() NodeMetadata {
	cp := make(NodeMetadata, len(nm))
	for k, v := range nm {
		cp[k] = v
	}
	return cp
}

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge is defined on the value-type, but node metadata map is
// itself a reference type, so if you want to maintain immutability, use copy.
func (nm NodeMetadata) Merge(other NodeMetadata) NodeMetadata {
	for k, v := range other {
		nm[k] = v // other takes precedence
	}
	return nm
}

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// RenderBy transforms a given Topology into a set of RenderableNodes, which
// the UI will render collectively as a graph. Note that a RenderableNode will
// always be rendered with other nodes, and therefore contains limited detail.
//
// RenderBy takes a a MapFunc, which defines how to group and label nodes. Npdes
// with the same mapped IDs will be merged.
func (t Topology) RenderBy(mapFunc MapFunc, pseudoFunc PseudoFunc) map[string]RenderableNode {
	nodes := map[string]RenderableNode{}

	// Build a set of RenderableNodes for all non-pseudo probes, and an
	// addressID to nodeID lookup map. Multiple addressIDs can map to the same
	// RenderableNodes.
	address2mapped := map[string]string{}
	for addressID, metadata := range t.NodeMetadatas {
		mapped, ok := mapFunc(addressID, metadata)
		if !ok {
			continue
		}

		// mapped.ID needs not be unique over all addressIDs. If not, we just overwrite
		// the existing data, on the assumption that the MapFunc returns the same
		// data.
		nodes[mapped.ID] = RenderableNode{
			ID:          mapped.ID,
			LabelMajor:  mapped.Major,
			LabelMinor:  mapped.Minor,
			Rank:        mapped.Rank,
			Pseudo:      false,
			Adjacency:   IDList{},            // later
			OriginHosts: IDList{},            // later
			OriginNodes: IDList{},            // later
			Metadata:    AggregateMetadata{}, // later
		}
		address2mapped[addressID] = mapped.ID
	}

	// Walk the graph and make connections.
	for src, dsts := range t.Adjacency {
		var (
			fields            = strings.SplitN(src, IDDelim, 2) // "<host>|<address>"
			srcOriginHostID   = fields[0]
			srcNodeAddress    = fields[1]
			srcRenderableID   = address2mapped[srcNodeAddress] // must exist
			srcRenderableNode = nodes[srcRenderableID]         // must exist
		)

		for _, dstNodeAddress := range dsts {
			dstRenderableID, ok := address2mapped[dstNodeAddress]
			if !ok {
				pseudoNode, ok := pseudoFunc(srcNodeAddress, srcRenderableNode, dstNodeAddress)
				if !ok {
					continue
				}
				dstRenderableID = pseudoNode.ID
				nodes[dstRenderableID] = RenderableNode{
					ID:         pseudoNode.ID,
					LabelMajor: pseudoNode.Major,
					LabelMinor: pseudoNode.Minor,
					Pseudo:     true,
					Metadata:   AggregateMetadata{}, // populated below - or not?
				}
				address2mapped[dstNodeAddress] = dstRenderableID
			}

			srcRenderableNode.Adjacency = srcRenderableNode.Adjacency.Add(dstRenderableID)
			srcRenderableNode.OriginHosts = srcRenderableNode.OriginHosts.Add(srcOriginHostID)
			srcRenderableNode.OriginNodes = srcRenderableNode.OriginNodes.Add(srcNodeAddress)
			edgeID := srcNodeAddress + IDDelim + dstNodeAddress
			if md, ok := t.EdgeMetadatas[edgeID]; ok {
				srcRenderableNode.Metadata.Merge(md.Transform())
			}
		}

		nodes[srcRenderableID] = srcRenderableNode
	}

	return nodes
}

// EdgeMetadata gives the metadata of an edge from the perspective of the
// srcRenderableID. Since an edgeID can have multiple edges on the address
// level, it uses the supplied mapping function to translate address IDs to
// renderable node (mapped) IDs.
func (t Topology) EdgeMetadata(mapFunc MapFunc, srcRenderableID, dstRenderableID string) EdgeMetadata {
	metadata := EdgeMetadata{}
	for edgeID, edgeMeta := range t.EdgeMetadatas {
		edgeParts := strings.SplitN(edgeID, IDDelim, 2)
		src := edgeParts[0]
		if src != TheInternet {
			mapped, _ := mapFunc(src, t.NodeMetadatas[src])
			src = mapped.ID
		}
		dst := edgeParts[1]
		if dst != TheInternet {
			mapped, _ := mapFunc(dst, t.NodeMetadatas[dst])
			dst = mapped.ID
		}
		if src == srcRenderableID && dst == dstRenderableID {
			metadata.Flatten(edgeMeta)
		}
	}
	return metadata
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
		} else if !reflect.DeepEqual(node, a[k]) {
			diff.Update = append(diff.Update, node)
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
