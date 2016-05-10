package report

import (
	"sort"
	"strings"

	"github.com/weaveworks/scope/common/mtime"
)

// MaxTableRows sets the limit on the table size to render
// TODO: this won't be needed once we send reports incrementally
const MaxTableRows = 20

// AddTable appends arbirary key-value pairs to the Node, returning a new node.
func (node Node) AddTable(prefix string, labels map[string]string) Node {
	count := 0
	for key, value := range labels {
		// It's enough to only include MaxTableRows+1
		// since they won't be rendered anyhow
		if count > MaxTableRows {
			break
		}
		node = node.WithLatest(prefix+key, mtime.Now(), value)
		count++

	}
	return node
}

// ExtractTable returns the key-value pairs with the given prefix from this Node,
func (node Node) ExtractTable(prefix string) (rows map[string]string, truncated bool) {
	rows = map[string]string{}
	truncated = false
	count := 0
	node.Latest.ForEach(func(key, value string) {
		if strings.HasPrefix(key, prefix) {
			if count >= MaxTableRows {
				truncated = true
				return
			}
			label := key[len(prefix):]
			rows[label] = value
			count++
		}
	})
	return rows, truncated
}

// Table is the type for a table in the UI.
type Table struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Rows         []MetadataRow `json:"rows"`
	WasTruncated bool          `json:"was_truncated"`
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

// TableTemplate describes how to render a table for the UI.
type TableTemplate struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Prefix string `json:"prefix"`
}

// Copy returns a value-copy of the TableTemplate
func (t TableTemplate) Copy() TableTemplate {
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

	return TableTemplate{
		ID:     max(t.ID, other.ID),
		Label:  max(t.Label, other.Label),
		Prefix: max(t.Prefix, other.Prefix),
	}
}

// TableTemplates is a mergeable set of TableTemplate
type TableTemplates map[string]TableTemplate

// Tables renders the TableTemplates for a given node.
func (t TableTemplates) Tables(node Node) []Table {
	var result []Table
	for _, template := range t {
		rows, truncated := node.ExtractTable(template.Prefix)
		table := Table{
			ID:           template.ID,
			Label:        template.Label,
			Rows:         []MetadataRow{},
			WasTruncated: truncated,
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
	result := t.Copy()
	for k, v := range other {
		if result == nil {
			result = TableTemplates{}
		}
		if existing, ok := result[k]; ok {
			v = v.Merge(existing)
		}
		result[k] = v
	}
	return result
}
