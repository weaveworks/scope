package report

// AggregateMetadata is a composable version of an EdgeMetadata. It's used
// when we want to merge nodes/edges for any reason.
//
// Even though we base it on EdgeMetadata, we can apply it to nodes, by
// summing up (merging) all of the {ingress, egress} metadatas of the
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

// Transform calculates a AggregateMetadata from an EdgeMetadata.
func (md EdgeMetadata) Transform() AggregateMetadata {
	m := AggregateMetadata{}
	if md.WithBytes {
		m[KeyBytesIngress] = int(md.BytesIngress)
		m[KeyBytesEgress] = int(md.BytesEgress)
	}
	if md.WithConnCountTCP {
		// The maximum is the maximum. No need to calculate anything.
		m[KeyMaxConnCountTCP] = int(md.MaxConnCountTCP)
	}
	return m
}

// Merge adds the fields from AggregateMetadata to r. r must be initialized.
func (r *AggregateMetadata) Merge(other AggregateMetadata) {
	for k, v := range other {
		(*r)[k] += v
	}
}
