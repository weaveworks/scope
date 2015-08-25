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

// MakeTopology gives you a Topology.
func MakeTopology() Topology {
	return Topology{
		Adjacency:     map[string]IDList{},
		EdgeMetadatas: map[string]EdgeMetadata{},
		NodeMetadatas: map[string]NodeMetadata{},
	}
}

// Copy returns a value copy of the Topology.
func (t Topology) Copy() Topology {
	return Topology{
		Adjacency:     t.Adjacency.Copy(),
		EdgeMetadatas: t.EdgeMetadatas.Copy(),
		NodeMetadatas: t.NodeMetadatas.Copy(),
	}
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (t Topology) Merge(other Topology) Topology {
	return Topology{
		Adjacency:     t.Adjacency.Merge(other.Adjacency),
		EdgeMetadatas: t.EdgeMetadatas.Merge(other.EdgeMetadatas),
		NodeMetadatas: t.NodeMetadatas.Merge(other.NodeMetadatas),
	}
}

// Adjacency is an adjacency-list encoding of the topology. Keys are adjacency
// IDs, values are lists of node IDs.
type Adjacency map[string]IDList

// Copy returns a value copy of the adjacency.
func (a Adjacency) Copy() Adjacency {
	cp := make(Adjacency, len(a))
	for k, v := range a {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (a Adjacency) Merge(other Adjacency) Adjacency {
	cp := a.Copy()
	for k, v := range other {
		cp[k] = cp[k].Merge(v)
	}
	return cp
}

// EdgeMetadatas collect metadata about each edge in a topology. Keys are a
// concatenation of node IDs.
type EdgeMetadatas map[string]EdgeMetadata

// Copy returns a value copy of the EdgeMetadatas.
func (e EdgeMetadatas) Copy() EdgeMetadatas {
	cp := make(EdgeMetadatas, len(e))
	for k, v := range e {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (e EdgeMetadatas) Merge(other EdgeMetadatas) EdgeMetadatas {
	cp := e.Copy()
	for k, v := range other {
		cp[k] = cp[k].Merge(v)
	}
	return cp
}

// NodeMetadatas collect metadata about each node in a topology. Keys are node
// IDs.
type NodeMetadatas map[string]NodeMetadata

// Copy returns a value copy of the NodeMetadatas.
func (n NodeMetadatas) Copy() NodeMetadatas {
	cp := make(NodeMetadatas, len(n))
	for k, v := range n {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (n NodeMetadatas) Merge(other NodeMetadatas) NodeMetadatas {
	cp := n.Copy()
	for k, v := range other {
		if _, ok := cp[k]; !ok { // don't overwrite
			cp[k] = v.Copy()
		}
	}
	return cp
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

// Copy returns a value copy of the EdgeMetadata.
func (e EdgeMetadata) Copy() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.EgressPacketCount),
		IngressPacketCount: cpu64ptr(e.IngressPacketCount),
		EgressByteCount:    cpu64ptr(e.EgressByteCount),
		IngressByteCount:   cpu64ptr(e.IngressByteCount),
		MaxConnCountTCP:    cpu64ptr(e.MaxConnCountTCP),
	}
}

func cpu64ptr(u *uint64) *uint64 {
	if u == nil {
		return nil
	}
	value := *u   // oh man
	return &value // this sucks
}

// Merge merges another EdgeMetadata into the receiver and returns the result.
// The receiver is not modified. The two edge metadatas should represent the
// same edge on different times.
func (e EdgeMetadata) Merge(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, max)
	return cp
}

// Flatten sums two EdgeMetadatas and returns the result. The receiver is not
// modified. The two edge metadata windows should be the same duration; they
// should represent different edges at the same time.
func (e EdgeMetadata) Flatten(other EdgeMetadata) EdgeMetadata {
	cp := e.Copy()
	cp.EgressPacketCount = merge(cp.EgressPacketCount, other.EgressPacketCount, sum)
	cp.IngressPacketCount = merge(cp.IngressPacketCount, other.IngressPacketCount, sum)
	cp.EgressByteCount = merge(cp.EgressByteCount, other.EgressByteCount, sum)
	cp.IngressByteCount = merge(cp.IngressByteCount, other.IngressByteCount, sum)
	// Note that summing of two maximums doesn't always give us the true
	// maximum. But it's a best effort.
	cp.MaxConnCountTCP = merge(cp.MaxConnCountTCP, other.MaxConnCountTCP, sum)
	return cp
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

// Copy returns a value copy of the NodeMetadata.
func (n NodeMetadata) Copy() NodeMetadata {
	cp := MakeNodeMetadata()
	for k, v := range n.Metadata {
		cp.Metadata[k] = v
	}
	for k, v := range n.Counters {
		cp.Counters[k] = v
	}
	return cp
}

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge does not modify the receiver.
func (n NodeMetadata) Merge(other NodeMetadata) NodeMetadata {
	cp := n.Copy()
	for k, v := range other.Metadata {
		cp.Metadata[k] = v // other takes precedence
	}
	for k, v := range other.Counters {
		cp.Counters[k] = n.Counters[k] + v
	}
	return cp
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
