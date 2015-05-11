package report

// AggregateMetadata is the per-second version of a EdgeMetadata. Rates,
// rather than counts. Only keys which are known are set, but they may be 0.
//
// Even though we base it on EdgeMetadata, we can apply it to nodes, by
// summing up (merging) all of the {ingress, egress} metadatas of the
// {incoming, outgoing} edges to the node.
type AggregateMetadata map[string]int

const (
	keyBytesIngress    = "ingress_bytes"
	keyBytesEgress     = "egress_bytes"
	keyMaxConnCountTCP = "max_conn_count_tcp"
)

// Render calculates a AggregateMetadata from an EdgeMetadata.
func (md EdgeMetadata) Render() AggregateMetadata {
	m := AggregateMetadata{}

	if md.WithBytes {
		m[keyBytesIngress] = int(md.BytesIngress)
		m[keyBytesEgress] = int(md.BytesEgress)
	}

	if md.WithConnCountTCP {
		// The maximum is the maximum. No need to calculate anything.
		m[keyMaxConnCountTCP] = int(md.MaxConnCountTCP)
	}

	return m
}

// Merge adds the fields from AggregateMetadata to r. r must be initialized.
func (r *AggregateMetadata) Merge(other AggregateMetadata) {
	for k, v := range other {
		(*r)[k] += v
	}
}
