package report

import (
	"reflect"
	"strconv"
)

// DetailedNode is given to the UI when a user clicks on a specific node. It
// contains detailed information about a node in the context of a rendered
// topology.
type DetailedNode struct {
	ID         string  `json:"id"`
	LabelMajor string  `json:"label_major"`
	LabelMinor string  `json:"label_minor,omitempty"`
	Pseudo     bool    `json:"pseudo,omitempty"`
	Tables     []Table `json:"tables"`
}

// Table is part of a detailed node.
type Table struct {
	Title   string `json:"title"`   // e.g. Bandwidth
	Numeric bool   `json:"numeric"` // should the major column be right-aligned?
	Rows    []Row  `json:"rows"`
}

// Row is part of a table.
type Row struct {
	Key        string `json:"key"`                   // e.g. Ingress
	ValueMajor string `json:"value_major"`           // e.g. 25
	ValueMinor string `json:"value_minor,omitempty"` // e.g. KB/s
}

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r Report, n RenderableNode) DetailedNode {
	tables := []Table{}
	{
		rows := []Row{}
		if val, ok := n.Metadata[KeyMaxConnCountTCP]; ok {
			rows = append(rows, Row{"TCP connections", strconv.FormatInt(int64(val), 10), ""})
		}
		if val, ok := n.Metadata[KeyBytesIngress]; ok {
			rows = append(rows, Row{"Bytes ingress", strconv.FormatInt(int64(val), 10), ""})
		}
		if val, ok := n.Metadata[KeyBytesEgress]; ok {
			rows = append(rows, Row{"Bytes egress", strconv.FormatInt(int64(val), 10), ""})
		}
		if len(rows) > 0 {
			tables = append(tables, Table{"Connections", true, rows})
		}
	}

	// RenderableNode may be the result of merge operation(s), and so may have
	// multiple origins. The ultimate goal here is to generate tables to view
	// in the UI, so we skip the intermediate representations, but we could
	// add them later.
outer:
	for _, id := range n.Origins {
		table, ok := r.OriginTable(id)
		if !ok {
			continue
		}
		// Naïve equivalence-based deduplication.
		for _, existing := range tables {
			if reflect.DeepEqual(existing, table) {
				continue outer
			}
		}
		tables = append(tables, table)
	}

	return DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
}
