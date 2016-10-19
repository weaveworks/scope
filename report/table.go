package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/scope/common/mtime"
)

// MaxTableRows sets the limit on the table size to render
// TODO: this won't be needed once we send reports incrementally
const (
	MaxTableRows          = 20
	TruncationCountPrefix = "table_truncation_count_"
)

// AddPrefixTable appends arbirary key-value pairs to the Node, returning a new node.
func (node Node) AddPrefixTable(prefix string, labels map[string]string) Node {
	count := 0
	for key, value := range labels {
		if count >= MaxTableRows {
			break
		}
		node = node.WithLatest(prefix+key, mtime.Now(), value)
		count++
	}
	if len(labels) > MaxTableRows {
		truncationCount := fmt.Sprintf("%d", len(labels)-MaxTableRows)
		node = node.WithLatest(TruncationCountPrefix+prefix, mtime.Now(), truncationCount)
	}
	return node
}

// extractTable returns the key-value pairs to build a table from this node
func (node Node) extractTable(template TableTemplate) (rows map[string]string, truncationCount int) {
	rows = map[string]string{}
	truncationCount = 0
	node.Latest.ForEach(func(key string, _ time.Time, value string) {
		if label, ok := template.FixedRows[key]; ok {
			rows[label] = value
		}
		if len(template.Prefix) > 0 && strings.HasPrefix(key, template.Prefix) {
			label := key[len(template.Prefix):]
			rows[label] = value
		}
	})
	if str, ok := node.Latest.Lookup(TruncationCountPrefix + template.Prefix); ok {
		if n, err := fmt.Sscanf(str, "%d", &truncationCount); n != 1 || err != nil {
			log.Warn("Unexpected truncation count format %q", str)
		}
	}
	return rows, truncationCount
}

// Table is the type for a table in the UI.
type Table struct {
	ID              string        `json:"id"`
	Label           string        `json:"label"`
	Rows            []MetadataRow `json:"rows"`
	TruncationCount int           `json:"truncationCount,omitempty"`
}

type tablesByID []Table

func (t tablesByID) Len() int           { return len(t) }
func (t tablesByID) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tablesByID) Less(i, j int) bool { return t[i].ID < t[j].ID }

// Copy returns a copy of the Table.
func (t Table) Copy() Table {
	result := Table{
		ID:    t.ID,
		Label: t.Label,
		Rows:  make([]MetadataRow, 0, len(t.Rows)),
	}
	for _, row := range t.Rows {
		result.Rows = append(result.Rows, row)
	}
	return result
}

// FixedRow describes a row which is part of a TableTemplate and whose value is extracted
// from a predetermined key
type FixedRow struct {
	Label string `json:"label"`
	Key   string `json:"key"`
}

// TableTemplate describes how to render a table for the UI.
type TableTemplate struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
	// FixedRows indicates what predetermined rows to render each entry is
	// indexed by the key to extract the row value is mapped to the row
	// label
	FixedRows map[string]string `json:"fixedRows"`
}

// Copy returns a value-copy of the TableTemplate
func (t TableTemplate) Copy() TableTemplate {
	fixedRowsCopy := make(map[string]string, len(t.FixedRows))
	for key, value := range t.FixedRows {
		fixedRowsCopy[key] = value
	}
	t.FixedRows = fixedRowsCopy
	return t
}

// Merge other into t, returning a fresh copy.  Does fieldwise max -
// whilst this isn't particularly meaningful, at least it idempotent,
// commutativite and associative.
func (t TableTemplate) Merge(other TableTemplate) TableTemplate {
	max := func(s1, s2 string) string {
		if s1 > s2 {
			return s1
		}
		return s2
	}

	fixedRows := t.FixedRows
	if len(other.FixedRows) > len(fixedRows) {
		fixedRows = other.FixedRows
	}

	return TableTemplate{
		ID:        max(t.ID, other.ID),
		Label:     max(t.Label, other.Label),
		Prefix:    max(t.Prefix, other.Prefix),
		FixedRows: fixedRows,
	}
}

// TableTemplates is a mergeable set of TableTemplate
type TableTemplates map[string]TableTemplate

// Tables renders the TableTemplates for a given node.
func (t TableTemplates) Tables(node Node) []Table {
	var result []Table
	for _, template := range t {
		rows, truncationCount := node.extractTable(template)
		table := Table{
			ID:              template.ID,
			Label:           template.Label,
			Rows:            []MetadataRow{},
			TruncationCount: truncationCount,
		}
		keys := make([]string, 0, len(rows))
		for k := range rows {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			table.Rows = append(table.Rows, MetadataRow{
				ID:    "label_" + key,
				Label: key,
				Value: rows[key],
			})
		}
		result = append(result, table)
	}
	sort.Sort(tablesByID(result))
	return result
}

// Copy returns a value copy of the TableTemplates
func (t TableTemplates) Copy() TableTemplates {
	if t == nil {
		return nil
	}
	result := TableTemplates{}
	for k, v := range t {
		result[k] = v.Copy()
	}
	return result
}

// Merge merges two sets of TableTemplates
func (t TableTemplates) Merge(other TableTemplates) TableTemplates {
	if t == nil && other == nil {
		return nil
	}
	result := make(TableTemplates, len(t))
	for k, v := range t {
		result[k] = v
	}
	for k, v := range other {
		if existing, ok := result[k]; ok {
			result[k] = v.Merge(existing)
		} else {
			result[k] = v
		}
	}
	return result
}
