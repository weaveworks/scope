package report

import (
	"fmt"
	"strings"
)

// Topology describes a specific view of a network. It consists of nodes and
// edges, represented by Adjacency, and metadata about those nodes and edges,
// represented by EdgeMetadatas and NodeMetadatas respectively.
type Topology struct {
	Adjacency
	EdgeMetadatas
	NodeMetadatas
}

// Merge merges another Topology into the receiver.
func (t *Topology) Merge(other Topology) {
	t.Adjacency.Merge(other.Adjacency)
	t.EdgeMetadatas.Merge(other.EdgeMetadatas)
	t.NodeMetadatas.Merge(other.NodeMetadatas)
}

// Adjacency is an adjacency-list encoding of the topology. Keys are node IDs,
// as produced by the relevant MappingFunc for the topology.
type Adjacency map[string]IDList

// Merge merges another Adjacency list into the receiver.
func (a *Adjacency) Merge(other Adjacency) {
	for addr, adj := range other {
		(*a)[addr] = (*a)[addr].Merge(adj)
	}
}

// EdgeMetadatas collect metadata about each edge in a topology. Keys are a
// concatenation of node IDs.
type EdgeMetadatas map[string]EdgeMetadata

// Merge merges another EdgeMetadatas into the receiver. If other is from
// another probe this is the union of both metadatas. Keys present in both are
// summed.
func (e *EdgeMetadatas) Merge(other EdgeMetadatas) {
	for id, edgemeta := range other {
		local := (*e)[id]
		local.Merge(edgemeta)
		(*e)[id] = local
	}
}

// NodeMetadatas collect metadata about each node in a topology. Keys are node
// IDs.
type NodeMetadatas map[string]NodeMetadata

// Merge merges another NodeMetadatas into the receiver.
func (m *NodeMetadatas) Merge(other NodeMetadatas) {
	for id, meta := range other {
		if _, ok := (*m)[id]; !ok {
			(*m)[id] = meta // not a copy
		}
	}
}

// EdgeMetadata describes a superset of the metadata that probes can possibly
// collect about a directed edge between two nodes in any topology.
type EdgeMetadata struct {
	EgressPacketCount  *uint64 `json:"egress_packet_count,omitempty"`
	IngressPacketCount *uint64 `json:"ingress_packet_count,omitempty"`
	EgressByteCount    *uint64 `json:"egress_byte_count,omitempty"`  // Transport layer
	IngressByteCount   *uint64 `json:"ingress_byte_count,omitempty"` // Transport layer
	MaxConnCountTCP    *uint64 `json:"max_conn_count_tcp,omitempty"`
}

// Merge merges another EdgeMetadata into the receiver. The two edge metadatas
// should represent the same edge on different times.
func (m *EdgeMetadata) Merge(other EdgeMetadata) {
	m.EgressPacketCount = merge(m.EgressPacketCount, other.EgressPacketCount, sum)
	m.IngressPacketCount = merge(m.IngressPacketCount, other.IngressPacketCount, sum)
	m.EgressByteCount = merge(m.EgressByteCount, other.EgressByteCount, sum)
	m.IngressByteCount = merge(m.IngressByteCount, other.IngressByteCount, sum)
	m.MaxConnCountTCP = merge(m.MaxConnCountTCP, other.MaxConnCountTCP, max)
}

// Flatten sums two EdgeMetadatas. Their windows should be the same duration;
// they should represent different edges at the same time.
func (m *EdgeMetadata) Flatten(other EdgeMetadata) {
	m.EgressPacketCount = merge(m.EgressPacketCount, other.EgressPacketCount, sum)
	m.IngressPacketCount = merge(m.IngressPacketCount, other.IngressPacketCount, sum)
	m.EgressByteCount = merge(m.EgressByteCount, other.EgressByteCount, sum)
	m.IngressByteCount = merge(m.IngressByteCount, other.IngressByteCount, sum)
	// Note that summing of two maximums doesn't always give us the true
	// maximum. But it's a best effort.
	m.MaxConnCountTCP = merge(m.MaxConnCountTCP, other.MaxConnCountTCP, sum)
}

// NodeMetadata describes a superset of the metadata that probes can collect
// about a given node in a given topology.
type NodeMetadata struct {
	Metadata map[string]string
	Counters map[string]int
}

// MakeNodeMetadata creates a new NodeMetadata with no initial metadata.
func MakeNodeMetadata() NodeMetadata {
	return MakeNodeMetadataWith(map[string]string{})
}

// MakeNodeMetadataWith creates a new NodeMetadata with the supplied map.
func MakeNodeMetadataWith(m map[string]string) NodeMetadata {
	return NodeMetadata{
		Metadata: m,
		Counters: map[string]int{},
	}
}

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge is defined on the value-type, but node metadata map is
// itself a reference type, so if you want to maintain immutability, use copy.
func (nm NodeMetadata) Merge(other NodeMetadata) NodeMetadata {
	for k, v := range other.Metadata {
		nm.Metadata[k] = v // other takes precedence
	}
	for k, v := range other.Counters {
		nm.Counters[k] = nm.Counters[k] + v
	}
	return nm
}

// Copy returns a value copy, useful for tests.
func (nm NodeMetadata) Copy() NodeMetadata {
	cp := MakeNodeMetadata()
	for k, v := range nm.Metadata {
		cp.Metadata[k] = v
	}
	for k, v := range nm.Counters {
		cp.Counters[k] = v
	}
	return cp
}

// NewTopology gives you a Topology.
func NewTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	// Check all edge metadata keys must have the appropriate entries in
	// adjacencies & node metadata.
	var errs []string
	for edgeID := range t.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := ParseEdgeID(edgeID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid edge ID %q", edgeID))
			continue
		}
		// For each edge, at least one of the ends must exist in nodemetadata
		if _, ok := t.NodeMetadatas[srcNodeID]; !ok {
			if _, ok := t.NodeMetadatas[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node metadatas missing for edge %q", edgeID))
			}
		}
		dstNodeIDs, ok := t.Adjacency[MakeAdjacencyID(srcNodeID)]
		if !ok {
			errs = append(errs, fmt.Sprintf("adjacency entries missing for source node ID %q (from edge %q)", srcNodeID, edgeID))
			continue
		}
		if !dstNodeIDs.Contains(dstNodeID) {
			errs = append(errs, fmt.Sprintf("adjacency destination missing for destination node ID %q (from edge %q)", dstNodeID, edgeID))
		}
	}

	// Check all adjancency keys has entries in NodeMetadata.
	for adjacencyID, dsts := range t.Adjacency {
		srcNodeID, ok := ParseAdjacencyID(adjacencyID)
		if !ok {
			errs = append(errs, fmt.Sprintf("invalid adjacency ID %q", adjacencyID))
			continue
		}
		for _, dstNodeID := range dsts {
			// For each edge, at least one of the ends must exist in nodemetadata
			if _, ok := t.NodeMetadatas[srcNodeID]; !ok {
				if _, ok := t.NodeMetadatas[dstNodeID]; !ok {
					errs = append(errs, fmt.Sprintf("node metadata missing from adjacency %q -> %q", srcNodeID, dstNodeID))
				}
			}
		}
	}

	// Check all node metadatas are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID := range t.NodeMetadatas {
		if t.NodeMetadatas[nodeID].Metadata == nil {
			errs = append(errs, fmt.Sprintf("node ID %q has nil metadata", nodeID))
		}
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s): %s", len(errs), strings.Join(errs, "; "))
	}

	return nil
}

func merge(dst, src *uint64, op func(uint64, uint64) uint64) *uint64 {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = new(uint64)
	}
	(*dst) = op(*dst, *src)
	return dst
}

func sum(dst, src uint64) uint64 {
	return dst + src
}

func max(dst, src uint64) uint64 {
	if dst > src {
		return dst
	}
	return src
}
