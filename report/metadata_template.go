package report

import (
	"sort"
	"strconv"
	"strings"
)

const (
	number = "number"
)

// FromLatest and friends denote the different fields where metadata can be
// gathered from.
const (
	FromLatest   = "latest"
	FromSets     = "sets"
	FromCounters = "counters"
)

// MetadataTemplate extracts some metadata rows from a node
type MetadataTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`    // Human-readable descriptor for this row
	Truncate int     `json:"truncate,omitempty"` // If > 0, truncate the value to this length.
	Datatype string  `json:"dataType,omitempty"`
	Priority float64 `json:"priority,omitempty"`
	From     string  `json:"from,omitempty"` // Defines how to get the value from a report node
}

// MetadataRow returns the row for a node
func (t MetadataTemplate) MetadataRow(n Node) (MetadataRow, bool) {
	from := fromDefault
	switch t.From {
	case FromLatest:
		from = fromLatest
	case FromSets:
		from = fromSets
	case FromCounters:
		from = fromCounters
	}
	if val, ok := from(n, t.ID); ok {
		return MetadataRow{
			ID:       t.ID,
			Label:    t.Label,
			Value:    val,
			Truncate: t.Truncate,
			Datatype: t.Datatype,
			Priority: t.Priority,
		}, true
	}
	return MetadataRow{}, false
}

func fromDefault(n Node, key string) (string, bool) {
	for _, from := range []func(n Node, key string) (string, bool){fromLatest, fromSets, fromCounters} {
		if val, ok := from(n, key); ok {
			return val, ok
		}
	}
	return "", false
}

func fromLatest(n Node, key string) (string, bool) {
	return n.Latest.Lookup(key)
}

func fromSets(n Node, key string) (string, bool) {
	val, ok := n.Sets.Lookup(key)
	return strings.Join(val, ", "), ok
}

func fromCounters(n Node, key string) (string, bool) {
	val := n.CountChildrenOfTopology(key)
	if val == 0 {
		return "", false
	}
	return strconv.Itoa(val), true
}

// MetadataRow is a row for the metadata table.
type MetadataRow struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Value    string  `json:"value"`
	Priority float64 `json:"priority,omitempty"`
	Datatype string  `json:"dataType,omitempty"`
	Truncate int     `json:"truncate,omitempty"`
}

// MetadataTemplates is a mergeable set of metadata templates
type MetadataTemplates map[string]MetadataTemplate

// MetadataRows returns the rows for a node
func (e MetadataTemplates) MetadataRows(n Node) []MetadataRow {
	if len(e) == 0 {
		return nil
	}
	rows := make([]MetadataRow, 0, len(e))
	for _, template := range e {
		if row, ok := template.MetadataRow(n); ok {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return nil
	}
	sort.Sort(MetadataRowsByPriority(rows))
	return rows
}

// Copy returns a value copy of the metadata templates
func (e MetadataTemplates) Copy() MetadataTemplates {
	if e == nil {
		return nil
	}
	result := make(MetadataTemplates, len(e))
	for k, v := range e {
		result[k] = v
	}
	return result
}

// Merge merges two sets of MetadataTemplates so far just ignores based
// on duplicate id key
func (e MetadataTemplates) Merge(other MetadataTemplates) MetadataTemplates {
	if len(other) > len(e) {
		e, other = other, e
	}
	if len(other) == 0 {
		return e
	}
	result := e.Copy()
	for k, v := range other {
		if existing, ok := result[k]; !ok || existing.Priority < v.Priority {
			result[k] = v
		}
	}
	return result
}

// MetadataRowsByPriority implements sort.Interface, so we can sort the rows by
// priority before rendering them to the UI.
type MetadataRowsByPriority []MetadataRow

// Len is part of sort.Interface.
func (m MetadataRowsByPriority) Len() int {
	return len(m)
}

// Swap is part of sort.Interface.
func (m MetadataRowsByPriority) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Less is part of sort.Interface.
func (m MetadataRowsByPriority) Less(i, j int) bool {
	return m[i].Priority < m[j].Priority
}
