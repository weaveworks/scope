package report

import (
	"fmt"
	"strings"
)

// Topology describes a specific view of a network. It consists of nodes and
// edges, and metadata about those nodes and edges, represented by
// EdgeMetadatas and Nodes respectively. Edges are directional, and embedded
// in the Node struct.
type Topology struct {
	Nodes // TODO(pb): remove Nodes intermediate type
}

// MakeTopology gives you a Topology.
func MakeTopology() Topology {
	return Topology{
		Nodes: map[string]Node{},
	}
}

// AddNode adds node to the topology under key nodeID; if a
// node already exists for this key, nmd is merged with that node.
// The same topology is returned to enable chaining.
// This method is different from all the other similar methods
// in that it mutates the Topology, to solve issues of GC pressure.
func (t Topology) AddNode(nodeID string, nmd Node) Topology {
	if existing, ok := t.Nodes[nodeID]; ok {
		nmd = nmd.Merge(existing)
	}
	t.Nodes[nodeID] = nmd
	return t
}

// Copy returns a value copy of the Topology.
func (t Topology) Copy() Topology {
	return Topology{
		Nodes: t.Nodes.Copy(),
	}
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (t Topology) Merge(other Topology) Topology {
	return Topology{
		Nodes: t.Nodes.Merge(other.Nodes),
	}
}

// Nodes is a collection of nodes in a topology. Keys are node IDs.
// TODO(pb): type Topology map[string]Node
type Nodes map[string]Node

// Copy returns a value copy of the Nodes.
func (n Nodes) Copy() Nodes {
	cp := make(Nodes, len(n))
	for k, v := range n {
		cp[k] = v.Copy()
	}
	return cp
}

// Merge merges the other object into this one, and returns the result object.
// The original is not modified.
func (n Nodes) Merge(other Nodes) Nodes {
	cp := n.Copy()
	for k, v := range other {
		if n, ok := cp[k]; ok { // don't overwrite
			v = v.Merge(n)
		}
		cp[k] = v
	}
	return cp
}

// Node describes a superset of the metadata that probes can collect about a
// given node in a given topology, along with the edges emanating from the
// node and metadata about those edges.
type Node struct {
	Metadata  `json:"metadata,omitempty"`
	Counters  `json:"counters,omitempty"`
	Adjacency IDList        `json:"adjacency"`
	Edges     EdgeMetadatas `json:"edges,omitempty"`
}

// MakeNode creates a new Node with no initial metadata.
func MakeNode() Node {
	return Node{
		Metadata:  Metadata{},
		Counters:  Counters{},
		Adjacency: MakeIDList(),
		Edges:     EdgeMetadatas{},
	}
}

// MakeNodeWith creates a new Node with the supplied map.
func MakeNodeWith(m map[string]string) Node {
	return MakeNode().WithMetadata(m)
}

// WithMetadata returns a fresh copy of n, with Metadata m merged in.
func (n Node) WithMetadata(m map[string]string) Node {
	result := n.Copy()
	result.Metadata = result.Metadata.Merge(m)
	return result
}

// WithCounters returns a fresh copy of n, with Counters c merged in.
func (n Node) WithCounters(c map[string]int) Node {
	result := n.Copy()
	result.Counters = result.Counters.Merge(c)
	return result
}

// WithAdjacent returns a fresh copy of n, with 'a' added to Adjacency
func (n Node) WithAdjacent(a string) Node {
	result := n.Copy()
	result.Adjacency = result.Adjacency.Add(a)
	return result
}

// WithEdge returns a fresh copy of n, with 'dst' added to Adjacency and md
// added to EdgeMetadata.
func (n Node) WithEdge(dst string, md EdgeMetadata) Node {
	result := n.Copy()
	result.Adjacency = result.Adjacency.Add(dst)
	result.Edges[dst] = md
	return result
}

// Copy returns a value copy of the Node.
func (n Node) Copy() Node {
	cp := MakeNode()
	cp.Metadata = n.Metadata.Copy()
	cp.Counters = n.Counters.Copy()
	cp.Adjacency = n.Adjacency.Copy()
	cp.Edges = n.Edges.Copy()
	return cp
}

// Merge mergses the individual components of a node and returns a
// fresh node.
func (n Node) Merge(other Node) Node {
	cp := n.Copy()
	cp.Metadata = cp.Metadata.Merge(other.Metadata)
	cp.Counters = cp.Counters.Merge(other.Counters)
	cp.Adjacency = cp.Adjacency.Merge(other.Adjacency)
	cp.Edges = cp.Edges.Merge(other.Edges)
	return cp
}

// Metadata is a string->string map.
type Metadata map[string]string

// Merge merges two node metadata maps together. In case of conflict, the
// other (right-hand) side wins. Always reassign the result of merge to the
// destination. Merge does not modify the receiver.
func (m Metadata) Merge(other Metadata) Metadata {
	result := m.Copy()
	for k, v := range other {
		result[k] = v // other takes precedence
	}
	return result
}

// Copy creates a deep copy of the Metadata.
func (m Metadata) Copy() Metadata {
	result := Metadata{}
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Counters is a string->int map.
type Counters map[string]int

// Merge merges two sets of counters into a fresh set of counters, summing
// values where appropriate.
func (c Counters) Merge(other Counters) Counters {
	result := c.Copy()
	for k, v := range other {
		result[k] = result[k] + v
	}
	return result
}

// Copy creates a deep copy of the Counters.
func (c Counters) Copy() Counters {
	result := Counters{}
	for k, v := range c {
		result[k] = v
	}
	return result
}

// EdgeMetadatas collect metadata about each edge in a topology. Keys are the
// remote node IDs, as in Adjacency.
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

// Flatten flattens all the EdgeMetadatas in this set and returns the result.
// The original is not modified.
func (e EdgeMetadatas) Flatten() EdgeMetadata {
	result := EdgeMetadata{}
	for _, v := range e {
		result = result.Flatten(v)
	}
	return result
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

// Reversed returns a value copy of the EdgeMetadata, with the direction reversed.
func (e EdgeMetadata) Reversed() EdgeMetadata {
	return EdgeMetadata{
		EgressPacketCount:  cpu64ptr(e.IngressPacketCount),
		IngressPacketCount: cpu64ptr(e.EgressPacketCount),
		EgressByteCount:    cpu64ptr(e.IngressByteCount),
		IngressByteCount:   cpu64ptr(e.EgressByteCount),
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

// Validate checks the topology for various inconsistencies.
func (t Topology) Validate() error {
	errs := []string{}

	// Check all node metadatas are valid, and the keys are parseable, i.e.
	// contain a scope.
	for nodeID, nmd := range t.Nodes {
		if nmd.Metadata == nil {
			errs = append(errs, fmt.Sprintf("node ID %q has nil metadata", nodeID))
		}
		if _, _, ok := ParseNodeID(nodeID); !ok {
			errs = append(errs, fmt.Sprintf("invalid node ID %q", nodeID))
		}

		// Check all adjancency keys has entries in Node.
		for _, dstNodeID := range nmd.Adjacency {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node metadata missing from adjacency %q -> %q", nodeID, dstNodeID))
			}
		}

		// Check all the edge metadatas have entries in adjacencies
		for dstNodeID := range nmd.Edges {
			if _, ok := t.Nodes[dstNodeID]; !ok {
				errs = append(errs, fmt.Sprintf("node %s metadatas missing for edge %q", dstNodeID, nodeID))
			}
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
