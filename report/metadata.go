package report

// RenderableMetadata is the per-second version of a EdgeMetadata. Only keys
// which are known are set, but they may be 0.
type RenderableMetadata map[string]int

const (
	keyBytesIngress    = "ingress_bytes"
	keyBytesEgress     = "egress_bytes"
	keyMaxConnCountTCP = "max_conn_count_tcp"
)

// Render calculates a RenderableMetadata from an EdgeMetadata.
func (md EdgeMetadata) Render() RenderableMetadata {
	m := RenderableMetadata{}

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

// Merge adds the fields from RenderableMetadata to r. r must be initialized.
func (r *RenderableMetadata) Merge(other RenderableMetadata) {
	for k, v := range other {
		(*r)[k] += v
	}
}
