package report

import (
	"math"
	"sort"
)

// MetricTemplate extracts a metric row from a node
type MetricTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`
	Format   string  `json:"format,omitempty"`
	Group    string  `json:"group,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

// MetricRow returns the row for a node
func (t MetricTemplate) MetricRow(n Node) (MetricRow, bool) {
	metric, ok := n.Metrics.Lookup(t.ID)
	if !ok {
		return MetricRow{}, false
	}
	row := MetricRow{
		ID:       t.ID,
		Label:    t.Label,
		Format:   t.Format,
		Group:    t.Group,
		Priority: t.Priority,
		Metric:   &metric,
	}
	if s, ok := metric.LastSample(); ok {
		row.Value = toFixed(s.Value, 2)
	}
	return row, true
}

// MetricTemplates is a mergeable set of metric templates
type MetricTemplates map[string]MetricTemplate

// MetricRows returns the rows for a node
func (e MetricTemplates) MetricRows(n Node) []MetricRow {
	if len(e) == 0 {
		return nil
	}
	rows := make([]MetricRow, 0, len(e))
	for _, template := range e {
		if row, ok := template.MetricRow(n); ok {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return nil
	}
	sort.Sort(MetricRowsByPriority(rows))
	return rows
}

// Copy returns a value copy of the metadata templates
func (e MetricTemplates) Copy() MetricTemplates {
	if e == nil {
		return nil
	}
	result := make(MetricTemplates, len(e))
	for k, v := range e {
		result[k] = v
	}
	return result
}

// Merge merges two sets of MetricTemplates so far just ignores based
// on duplicate id key
func (e MetricTemplates) Merge(other MetricTemplates) MetricTemplates {
	if e == nil && other == nil {
		return nil
	}
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

// toFixed truncates decimals of float64 down to specified precision
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int64(num*output)) / output
}
