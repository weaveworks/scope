package report

// Merge functions for all topology datatypes. The general semantics are that
// the receiver is modified, and what's merged in isn't.

// Merge merges another Report into the receiver. Pass addWindows true if the
// reports represent distinct (non-overlapping) periods of time.
func (r *Report) Merge(other Report) {
	r.Endpoint.Merge(other.Endpoint)
	r.Address.Merge(other.Address)
	r.Process.Merge(other.Process)
	r.Container.Merge(other.Container)
	r.ContainerImage.Merge(other.ContainerImage)
	r.Host.Merge(other.Host)
	r.Overlay.Merge(other.Overlay)
	r.Sampling.Merge(other.Sampling)
	r.Window += other.Window
}

// Merge merges another Topology into the receiver.
func (t *Topology) Merge(other Topology) {
	t.Adjacency.Merge(other.Adjacency)
	t.EdgeMetadatas.Merge(other.EdgeMetadatas)
	t.NodeMetadatas.Merge(other.NodeMetadatas)
}

// Merge merges another Adjacency list into the receiver.
func (a *Adjacency) Merge(other Adjacency) {
	for addr, adj := range other {
		(*a)[addr] = (*a)[addr].Merge(adj)
	}
}

// Merge merges another NodeMetadatas into the receiver.
func (m *NodeMetadatas) Merge(other NodeMetadatas) {
	for id, meta := range other {
		if _, ok := (*m)[id]; !ok {
			(*m)[id] = meta // not a copy
		}
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
	return nm
}

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

// Merge combines two sampling structures via simple addition.
func (s *Sampling) Merge(other Sampling) {
	s.Count += other.Count
	s.Total += other.Total
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
