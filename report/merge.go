package report

// Merge functions for all topology datatypes. The general semantics are that
// the receiver is modified, and what's merged in isn't.

// Merge merges another Report into the receiver.
func (r *Report) Merge(other Report) {
	r.Endpoint.Merge(other.Endpoint)
	r.Address.Merge(other.Address)
	r.Process.Merge(other.Process)
	r.Container.Merge(other.Container)
	r.ContainerImage.Merge(other.ContainerImage)
	r.Host.Merge(other.Host)
	r.Overlay.Merge(other.Overlay)
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
		(*a)[addr] = (*a)[addr].Add(adj...)
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
	if other.WithBytes {
		m.WithBytes = true
		m.BytesIngress += other.BytesIngress
		m.BytesEgress += other.BytesEgress
	}
	if other.WithConnCountTCP {
		m.WithConnCountTCP = true
		if other.MaxConnCountTCP > m.MaxConnCountTCP {
			m.MaxConnCountTCP = other.MaxConnCountTCP
		}
	}
}

// Flatten sums two EdgeMetadatas. Their windows should be the same duration;
// they should represent different edges at the same time.
func (m *EdgeMetadata) Flatten(other EdgeMetadata) {
	if other.WithBytes {
		m.WithBytes = true
		m.BytesIngress += other.BytesIngress
		m.BytesEgress += other.BytesEgress
	}
	if other.WithConnCountTCP {
		m.WithConnCountTCP = true
		// Note: summing of two maximums doesn't always give the true maximum.
		// But it's our Best Effort effort.
		m.MaxConnCountTCP += other.MaxConnCountTCP
	}
}
