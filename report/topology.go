package report

import "net"

// Topology represents a directed graph of nodes and edges. This is the core
// data type, made of directly observed (via the probe) nodes, and their
// connections. Not to be confused with an API topology, which is made of
// renderable nodes, and produced by passing a report (i.e. multiple
// topologies) through a rendering transformation.
type Topology struct {
	Adjacency
	NodeMetadatas
	EdgeMetadatas
}

// MakeTopology produces a new topology, ready for use. It's the only correct
// way to produce topologies for general use, so please use it.
func MakeTopology() Topology {
	return Topology{
		Adjacency:     Adjacency{},
		NodeMetadatas: NodeMetadatas{},
		EdgeMetadatas: EdgeMetadatas{},
	}
}

// Copy returns a value copy, useful for tests.
func (t Topology) Copy() Topology {
	return Topology{
		Adjacency:     t.Adjacency.Copy(),
		NodeMetadatas: t.NodeMetadatas.Copy(),
		EdgeMetadatas: t.EdgeMetadatas.Copy(),
	}
}

// Merge merges two topologies together, returning the result. Always reassign
// the result of merge to the destination. Merge is defined on topology as a
// value-type, but topology contains reference fields, so if you want to
// maintain immutability, use copy.
func (t Topology) Merge(other Topology) Topology {
	t.Adjacency = t.Adjacency.Merge(other.Adjacency)
	t.NodeMetadatas = t.NodeMetadatas.Merge(other.NodeMetadatas)
	t.EdgeMetadatas = t.EdgeMetadatas.Merge(other.EdgeMetadatas)
	return t
}

// Squash squashes all non-local nodes in the topology to a super-node called
// the Internet.
func (t Topology) Squash(f IDAddresser, localNets []*net.IPNet) Topology {
	isRemote := func(ip net.IP) bool { return !netsContain(localNets, ip) }
	for srcID, dstIDs := range t.Adjacency {
		newDstIDs := make(IDList, 0, len(dstIDs))
		for _, dstID := range dstIDs {
			if ip := f(dstID); ip != nil && isRemote(ip) {
				dstID = TheInternet
			}
			newDstIDs = newDstIDs.Add(dstID)
		}
		t.Adjacency[srcID] = newDstIDs
	}
	return t
}

func netsContain(nets []*net.IPNet, ip net.IP) bool {
	for _, net := range nets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// Adjacency represents an adjacency list.
type Adjacency map[string]IDList

// Copy returns a value copy, useful for tests.
func (a Adjacency) Copy() Adjacency {
	cp := make(Adjacency, len(a))
	for id, idList := range a {
		cp[id] = idList.Copy()
	}
	return cp
}

// Merge merges two adjacencies together, returning the union set of IDs.
// Always reassign the result of merge to the destination. Merge is defined
// on adjacency as a value-type, but adjacency is itself a reference type, so
// if you want to maintain immutability, use copy.
func (a Adjacency) Merge(other Adjacency) Adjacency {
	for id, idList := range other {
		a[id] = a[id].Add(idList...)
	}
	return a
}

// NodeMetadatas represents multiple node metadatas, keyed by node ID.
type NodeMetadatas map[string]NodeMetadata

// Copy returns a value copy, useful for tests.
func (nms NodeMetadatas) Copy() NodeMetadatas {
	cp := make(NodeMetadatas, len(nms))
	for id, nm := range nms {
		cp[id] = nm.Copy()
	}
	return cp
}

// Merge merges two node metadata collections together, returning a semantic
// union set of metadatas. In the cases where keys conflict within an
// individual node metadata map, the other (right-hand) side wins. Always
// reassign the result of merge to the destination. Merge is defined on the
// value-type, but node metadata collection is itself a reference type, so if
// you want to maintain immutability, use copy.
func (nms NodeMetadatas) Merge(other NodeMetadatas) NodeMetadatas {
	for id, md := range other {
		if _, ok := nms[id]; !ok {
			nms[id] = NodeMetadata{}
		}
		nms[id] = nms[id].Merge(md)
	}
	return nms
}

// EdgeMetadatas represents multiple edge metadatas, keyed by edge ID.
type EdgeMetadatas map[string]EdgeMetadata

// Copy returns a value copy, useful for tests.
func (ems EdgeMetadatas) Copy() EdgeMetadatas {
	cp := make(EdgeMetadatas, len(ems))
	for id, em := range ems {
		cp[id] = em.Copy()
	}
	return cp
}

// Merge merges two edge metadata collections together, returning a logical
// union set of metadatas. Always reassign the result of merge to the
// destination. Merge is defined on the value-type, but edge metadata
// collection is itself a reference type, so if you want to maintain
// immutability, use copy.
func (ems EdgeMetadatas) Merge(other EdgeMetadatas) EdgeMetadatas {
	for id, md := range other {
		if _, ok := ems[id]; !ok {
			ems[id] = EdgeMetadata{}
		}
		ems[id] = ems[id].Merge(md)
	}
	return ems
}

// NodeMetadata is a simple string-to-string map.
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

// GetDefault returns the value for key, or def if key is not defined.
func (nm NodeMetadata) GetDefault(key, def string) string {
	val, ok := nm[key]
	if !ok {
		return def
	}
	return val
}

// EdgeMetadata represents aggregatable information about a specific edge
// between two nodes. EdgeMetadata is frequently merged; be careful to think
// about merge semantics when modifying this structure.
type EdgeMetadata struct {
	WithBytes    bool `json:"with_bytes,omitempty"`
	BytesEgress  uint `json:"bytes_egress,omitempty"`  // src -> dst
	BytesIngress uint `json:"bytes_ingress,omitempty"` // src <- dst

	WithConnCountTCP bool `json:"with_conn_count_tcp,omitempty"`
	MaxConnCountTCP  uint `json:"max_conn_count_tcp,omitempty"`
}

// Copy returns a value copy, useful for tests. It actually doesn't do
// anything here, since EdgeMetadata has no reference fields. But we keep it,
// for API consistency.
func (em EdgeMetadata) Copy() EdgeMetadata {
	return em
}

// Merge merges two edge metadata structs together. This is an important
// operation, so please think carefully when adding things here. Always
// reassign the result of merge to the destination, as merge is defined on a
// value-type, and there are no reference types here.
func (em EdgeMetadata) Merge(other EdgeMetadata) EdgeMetadata {
	if other.WithBytes {
		em.WithBytes = true
		em.BytesIngress += other.BytesIngress
		em.BytesEgress += other.BytesEgress
	}
	if other.WithConnCountTCP {
		em.WithConnCountTCP = true
		if other.MaxConnCountTCP > em.MaxConnCountTCP {
			em.MaxConnCountTCP = other.MaxConnCountTCP
		}
	}
	return em
}

// Flatten sums two EdgeMetadatas. They must represent the same window of
// time, i.e. both EdgeMetadatas should be from the same report.
func (em EdgeMetadata) Flatten(other EdgeMetadata) EdgeMetadata {
	if other.WithBytes {
		em.WithBytes = true
		em.BytesIngress += other.BytesIngress
		em.BytesEgress += other.BytesEgress
	}
	if other.WithConnCountTCP {
		em.WithConnCountTCP = true
		// Note that summing of two maximums doesn't always give the true
		// maximum. But it's our Best-Effort effort.
		em.MaxConnCountTCP += other.MaxConnCountTCP
	}
	return em
}

// Export transforms an EdgeMetadata to an AggregateMetadata.
func (em EdgeMetadata) Export() AggregateMetadata {
	amd := AggregateMetadata{}
	if em.WithBytes {
		amd[KeyBytesIngress] = int(em.BytesIngress)
		amd[KeyBytesEgress] = int(em.BytesEgress)
	}
	if em.WithConnCountTCP {
		// The maximum is the maximum. No need to calculate anything.
		amd[KeyMaxConnCountTCP] = int(em.MaxConnCountTCP)
	}
	return amd
}

// AggregateMetadata is a composable version of an EdgeMetadata. It's used
// when we want to merge nodes/edges for any reason.
//
// It takes its data from EdgeMetadatas, but we can apply it to nodes, by
// summing up (flattening) all of the {ingress, egress} metadatas of the
// {incoming, outgoing} edges to the node.
type AggregateMetadata map[string]int

const (
	// KeyBytesIngress is the aggregate metadata key for the total count of
	// ingress bytes.
	KeyBytesIngress = "ingress_bytes"
	// KeyBytesEgress is the aggregate metadata key for the total count of
	// egress bytes.
	KeyBytesEgress = "egress_bytes"
	// KeyMaxConnCountTCP is the aggregate metadata key for the maximum number
	// of simultaneous observed TCP connections in the window.
	KeyMaxConnCountTCP = "max_conn_count_tcp"
)

// Copy returns a value copy, useful for tests.
func (amd AggregateMetadata) Copy() AggregateMetadata {
	cp := make(AggregateMetadata, len(amd))
	for k, v := range amd {
		cp[k] = v
	}
	return cp
}

// Merge merges two aggregate metadatas together. Always reassign the result
// of merge to the destination. Merge is defined on the value-type, but
// aggregate metadata is itself a reference type, so if you want to maintain
// immutability, use copy.
func (amd AggregateMetadata) Merge(other AggregateMetadata) AggregateMetadata {
	for k, v := range other {
		amd[k] += v
	}
	return amd
}
