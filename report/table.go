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
	MaxTableRows           = 20
	TableEntryKeySeparator = "___"
	TruncationCountPrefix  = "table_truncation_count_"
	MulticolumnTableType   = "multicolumn-table"
	PropertyListType       = "property-list"
)

// withTableTruncationInformation appends table truncation info to the node, returning the new node.
func (node Node) withTableTruncationInformation(prefix string, totalRowsCount int, now time.Time) Node {
	if totalRowsCount > MaxTableRows {
		truncationCount := fmt.Sprintf("%d", totalRowsCount-MaxTableRows)
		node = node.WithLatest(TruncationCountPrefix+prefix, now, truncationCount)
	}
	return node
}

// AddPrefixMulticolumnTable appends arbitrary rows to the Node, returning a new node.
func (node Node) AddPrefixMulticolumnTable(prefix string, rows []Row) Node {
	now := mtime.Now()
	addedRowsCount := 0
	for _, row := range rows {
		if addedRowsCount >= MaxTableRows {
			break
		}
		// Add all the row values as separate entries
		for columnID, value := range row.Entries {
			key := strings.Join([]string{row.ID, columnID}, TableEntryKeySeparator)
			node = node.WithLatest(prefix+key, now, value)
		}
		addedRowsCount++
	}
	return node.withTableTruncationInformation(prefix, len(rows), now)
}

// AddPrefixPropertyList appends arbitrary key-value pairs to the Node, returning a new node.
func (node Node) AddPrefixPropertyList(prefix string, propertyList map[string]string) Node {
	now := mtime.Now()
	addedPropertiesCount := 0
	for label, value := range propertyList {
		if addedPropertiesCount >= MaxTableRows {
			break
		}
		node = node.WithLatest(prefix+label, now, value)
		addedPropertiesCount++
	}
	return node.withTableTruncationInformation(prefix, len(propertyList), now)
}

// WithoutPrefix returns the string with trimmed prefix and a
// boolean information of whether that prefix was really there.
// NOTE: Consider moving this function to utilities.
func WithoutPrefix(s string, prefix string) (string, bool) {
	return strings.TrimPrefix(s, prefix), len(prefix) > 0 && strings.HasPrefix(s, prefix)
}

// ExtractMulticolumnTable returns the rows to build a multicolumn table from this node
func (node Node) ExtractMulticolumnTable(template TableTemplate) (rows []Row) {
	rowsMapByID := map[string]Row{}

	// Itearate through the whole of our map to extract all the values with the key
	// with the given prefix. Since multicolumn tables don't support fixed rows (yet),
	// all the table values will be stored under the table prefix.
	// NOTE: It would be nice to optimize this part by only iterating through the keys
	// with the given prefix. If it is possible to traverse the keys in the Latest map
	// in a sorted order, then having LowerBoundEntry(key) and UpperBoundEntry(key)
	// methods should be enough to implement ForEachWithPrefix(prefix) straightforwardly.
	node.Latest.ForEach(func(key string, _ time.Time, value string) {
		if keyWithoutPrefix, ok := WithoutPrefix(key, template.Prefix); ok {
			ids := strings.Split(keyWithoutPrefix, TableEntryKeySeparator)
			rowID, columnID := ids[0], ids[1]
			// If the row with the given ID doesn't yet exist, we create an empty one.
			if _, ok := rowsMapByID[rowID]; !ok {
				rowsMapByID[rowID] = Row{
					ID:      rowID,
					Entries: map[string]string{},
				}
			}
			// At this point, the row with that ID always exists, so we just update the value.
			rowsMapByID[rowID].Entries[columnID] = value
		}
	})

	// Gather a list of rows.
	rows = make([]Row, 0, len(rowsMapByID))
	for _, row := range rowsMapByID {
		rows = append(rows, row)
	}

	// Return the rows sorted by ID.
	sort.Sort(rowsByID(rows))
	return rows
}

// ExtractPropertyList returns the rows to build a property list from this node
func (node Node) ExtractPropertyList(template TableTemplate) (rows []Row) {
	valuesMapByLabel := map[string]string{}

	// Itearate through the whole of our map to extract all the values with the key
	// with the given prefix as well as the keys corresponding to the fixed table rows.
	node.Latest.ForEach(func(key string, _ time.Time, value string) {
		if label, ok := template.FixedRows[key]; ok {
			valuesMapByLabel[label] = value
		} else if label, ok := WithoutPrefix(key, template.Prefix); ok {
			valuesMapByLabel[label] = value
		}
	})

	// Gather a label-value formatted list of rows.
	rows = make([]Row, 0, len(valuesMapByLabel))
	for label, value := range valuesMapByLabel {
		rows = append(rows, Row{
			ID: "label_" + label,
			Entries: map[string]string{
				"label": label,
				"value": value,
			},
		})
	}

	// Return the rows sorted by ID.
	sort.Sort(rowsByID(rows))
	return rows
}

// ExtractTable returns the rows to build either a property list or a generic table from this node
func (node Node) ExtractTable(template TableTemplate) (rows []Row, truncationCount int) {
	switch template.Type {
	case MulticolumnTableType:
		rows = node.ExtractMulticolumnTable(template)
	default: // By default assume it's a property list (for backward compatibility).
		rows = node.ExtractPropertyList(template)
	}

	truncationCount = 0
	if str, ok := node.Latest.Lookup(TruncationCountPrefix + template.Prefix); ok {
		if n, err := fmt.Sscanf(str, "%d", &truncationCount); n != 1 || err != nil {
			log.Warn("Unexpected truncation count format %q", str)
		}
	}

	return rows, truncationCount
}

// Column is the type for multi-column tables in the UI.
type Column struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	DataType string `json:"dataType"`
}

// Row is the type that holds the table data for the UI. Entries map from column ID to cell value.
type Row struct {
	ID      string            `json:"id"`
	Entries map[string]string `json:"entries"`
}

type rowsByID []Row

func (t rowsByID) Len() int           { return len(t) }
func (t rowsByID) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t rowsByID) Less(i, j int) bool { return t[i].ID < t[j].ID }

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

	// NOTE: Consider actually merging the columns and fixed rows.
	fixedRows := t.FixedRows
	if len(other.FixedRows) > len(fixedRows) {
		fixedRows = other.FixedRows
	}
	columns := t.Columns
	if len(other.Columns) > len(columns) {
		columns = other.Columns
	}

	// TODO: Refactor the merging logic, as mixing the types now might result in
	// invalid tables. Maybe we should return an error if the types are different?
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
		// Extract the type from the template; default to
		// property list for backwards-compatibility.
		tableType := template.Type
		if tableType == "" {
			tableType = PropertyListType
		}
		result = append(result, Table{
			ID:              template.ID,
			Label:           template.Label,
			Columns:         template.Columns,
			Type:            tableType,
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
