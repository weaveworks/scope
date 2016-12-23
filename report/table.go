package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/weaveworks/common/mtime"
)

// MaxTableRows sets the limit on the table size to render
// TODO: this won't be needed once we send reports incrementally
const (
	MaxTableRows          = 20
	TruncationCountPrefix = "table_truncation_count_"
	MulticolumnTableType  = "multicolumn-table"
	PropertyListType      = "property-list"
)

// AddPrefixTable appends arbitrary rows to the Node, returning a new node.
func (node Node) AddPrefixTable(prefix string, rows []Row) Node {
	count := 0
	for _, row := range rows {
		if count >= MaxTableRows {
			break
		}
		// TODO: Figure a more natural way of storing rows
		for column, value := range row.Entries {
			key := fmt.Sprintf("%s %s", row.ID, column)
			node = node.WithLatest(prefix+key, mtime.Now(), value)
		}
		count++
	}
	if len(rows) > MaxTableRows {
		truncationCount := fmt.Sprintf("%d", len(rows)-MaxTableRows)
		node = node.WithLatest(TruncationCountPrefix+prefix, mtime.Now(), truncationCount)
	}
	return node
}

// AddPrefixLabels appends arbitrary key-value pairs to the Node, returning a new node.
func (node Node) AddPrefixLabels(prefix string, labels map[string]string) Node {
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

// ExtractTable returns the rows to build a table from this node
func (node Node) ExtractTable(template TableTemplate) (rows []Row, truncationCount int) {
	rows = []Row{}
	switch template.Type {
	case MulticolumnTableType:
		keyRows := map[string]Row{}
		node.Latest.ForEach(func(key string, _ time.Time, value string) {
			if len(template.Prefix) > 0 && strings.HasPrefix(key, template.Prefix) {
				rowID, column := "", ""
				fmt.Sscanf(key[len(template.Prefix):], "%s %s", &rowID, &column)
				if _, ok := keyRows[rowID]; !ok {
					keyRows[rowID] = Row{
						ID:      rowID,
						Entries: map[string]string{},
					}
				}
				keyRows[rowID].Entries[column] = value
			}
		})
		for _, row := range keyRows {
			rows = append(rows, row)
		}
	// By default assume it's a property list (for backward compatibility)
	default:
		keyValues := map[string]string{}
		node.Latest.ForEach(func(key string, _ time.Time, value string) {
			if label, ok := template.FixedRows[key]; ok {
				keyValues[label] = value
			}
			if len(template.Prefix) > 0 && strings.HasPrefix(key, template.Prefix) {
				label := key[len(template.Prefix):]
				keyValues[label] = value
			}
		})
		labels := make([]string, 0, len(rows))
		for label := range keyValues {
			labels = append(labels, label)
		}
		sort.Strings(labels)
		for _, label := range labels {
			rows = append(rows, Row{
				ID: "label_" + label,
				Entries: map[string]string{
					"label": label,
					"value": keyValues[label],
				},
			})
		}
	}

	truncationCount = 0
	if str, ok := node.Latest.Lookup(TruncationCountPrefix + template.Prefix); ok {
		if n, err := fmt.Sscanf(str, "%d", &truncationCount); n != 1 || err != nil {
			log.Warn("Unexpected truncation count format %q", str)
		}
	}

	return rows, truncationCount
}

type Column struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	DataType string `json:"dataType"`
}

type Row struct {
	ID      string            `json:"id"`
	Entries map[string]string `json:"entries"`
}

// Copy returns a copy of the Row.
func (r Row) Copy() Row {
	entriesCopy := make(map[string]string, len(r.Entries))
	for key, value := range r.Entries {
		entriesCopy[key] = value
	}
	r.Entries = entriesCopy
	return r
}

// Table is the type for a table in the UI.
type Table struct {
	ID              string   `json:"id"`
	Label           string   `json:"label"`
	Type            string   `json:"type"`
	Columns         []Column `json:"columns"`
	Rows            []Row    `json:"rows"`
	TruncationCount int      `json:"truncationCount,omitempty"`
}

type tablesByID []Table

func (t tablesByID) Len() int           { return len(t) }
func (t tablesByID) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t tablesByID) Less(i, j int) bool { return t[i].ID < t[j].ID }

// Copy returns a copy of the Table.
func (t Table) Copy() Table {
	result := Table{
		ID:      t.ID,
		Label:   t.Label,
		Type:    t.Type,
		Columns: make([]Column, 0, len(t.Columns)),
		Rows:    make([]Row, 0, len(t.Rows)),
	}
	for _, column := range t.Columns {
		result.Columns = append(result.Columns, column)
	}
	for _, row := range t.Rows {
		result.Rows = append(result.Rows, row)
	}
	return result
}

// TableTemplate describes how to render a table for the UI.
type TableTemplate struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Prefix  string   `json:"prefix"`
	Type    string   `json:"type"`
	Columns []Column `json:"columns"`
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

	columns := t.Columns
	if len(other.Columns) > len(columns) {
		columns = other.Columns
	}

	// TODO: Refactor the merging logic, as mixing
	// the types now might result in invalid tables.
	return TableTemplate{
		ID:        max(t.ID, other.ID),
		Label:     max(t.Label, other.Label),
		Prefix:    max(t.Prefix, other.Prefix),
		Type:      max(t.Type, other.Type),
		Columns:   columns,
		FixedRows: fixedRows,
	}
}

// TableTemplates is a mergeable set of TableTemplate
type TableTemplates map[string]TableTemplate

// Tables renders the TableTemplates for a given node.
func (t TableTemplates) Tables(node Node) []Table {
	var result []Table
	for _, template := range t {
		rows, truncationCount := node.ExtractTable(template)
		result = append(result, Table{
			ID:              template.ID,
			Label:           template.Label,
			Type:            template.Type,
			Columns:         template.Columns,
			Rows:            rows,
			TruncationCount: truncationCount,
		})
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
