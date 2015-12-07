package detailed

import (
	"encoding/json"
	"math"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	defaultFormat  = ""
	filesizeFormat = "filesize"
	percentFormat  = "percent"
)

// MetricRow is a tuple of data used to render a metric as a sparkline and
// accoutrements.
type MetricRow struct {
	ID     string
	Label  string
	Format string
	Group  string
	Value  float64
	Metric *report.Metric
}

// Copy returns a value copy of the MetricRow
func (m MetricRow) Copy() MetricRow {
	metric := m.Metric.Copy()
	return MetricRow{
		ID:     m.ID,
		Label:  m.Label,
		Format: m.Format,
		Group:  m.Group,
		Value:  m.Value,
		Metric: &metric,
	}
}

// MarshalJSON marshals this MetricRow to json. It takes the basic Metric
// rendering, then adds some row-specific fields.
func (m MetricRow) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID     string  `json:"id"`
		Label  string  `json:"label"`
		Format string  `json:"format,omitempty"`
		Group  string  `json:"group,omitempty"`
		Value  float64 `json:"value"`
		report.WireMetrics
	}{
		ID:          m.ID,
		Label:       m.Label,
		Format:      m.Format,
		Group:       m.Group,
		Value:       m.Value,
		WireMetrics: m.Metric.ToIntermediate(),
	})
}

func metricRow(id, label string, metric report.Metric, format, group string) MetricRow {
	var last float64
	if s := metric.LastSample(); s != nil {
		last = s.Value
	}
	return MetricRow{
		ID:     id,
		Label:  label,
		Format: format,
		Group:  group,
		Value:  toFixed(last, 2),
		Metric: &metric,
	}
}

// toFixed truncates decimals of float64 down to specified precision
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int64(num*output)) / output
}

// NodeMetrics produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetrics(n report.Node) []MetricRow {
	renderers := map[string]func(report.Node) []MetricRow{
		"process":   processNodeMetrics,
		"container": containerNodeMetrics,
		"host":      hostNodeMetrics,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		return renderer(n)
	}
	return nil
}

func processNodeMetrics(nmd report.Node) []MetricRow {
	rows := []MetricRow{}
	for _, tuple := range []struct {
		ID, Label, fmt string
	}{
		{process.CPUUsage, "CPU Usage", percentFormat},
		{process.MemoryUsage, "Memory Usage", filesizeFormat},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			rows = append(rows, metricRow(
				tuple.ID,
				tuple.Label,
				val,
				tuple.fmt,
				"",
			))
		}
	}
	return rows
}

func containerNodeMetrics(nmd report.Node) []MetricRow {
	rows := []MetricRow{}
	if val, ok := nmd.Metrics[docker.CPUTotalUsage]; ok {
		rows = append(rows, metricRow(
			docker.CPUTotalUsage,
			"CPU Usage",
			val,
			percentFormat,
			"",
		))
	}
	if val, ok := nmd.Metrics[docker.MemoryUsage]; ok {
		rows = append(rows, metricRow(
			docker.MemoryUsage,
			"Memory Usage",
			val,
			filesizeFormat,
			"",
		))
	}
	return rows
}

func hostNodeMetrics(nmd report.Node) []MetricRow {
	// Ensure that all metrics have the same max
	maxLoad := 0.0
	for _, id := range []string{host.Load1, host.Load5, host.Load15} {
		if metric, ok := nmd.Metrics[id]; ok {
			if metric.Len() == 0 {
				continue
			}
			if metric.Max > maxLoad {
				maxLoad = metric.Max
			}
		}
	}

	rows := []MetricRow{}
	for _, tuple := range []struct{ ID, Label, fmt string }{
		{host.CPUUsage, "CPU Usage", percentFormat},
		{host.MemUsage, "Memory Usage", filesizeFormat},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			rows = append(rows, metricRow(tuple.ID, tuple.Label, val, tuple.fmt, ""))
		}
	}
	for _, tuple := range []struct{ ID, Label string }{
		{host.Load1, "Load (1m)"},
		{host.Load5, "Load (5m)"},
		{host.Load15, "Load (15m)"},
	} {
		if val, ok := nmd.Metrics[tuple.ID]; ok {
			val.Max = maxLoad
			rows = append(rows, metricRow(tuple.ID, tuple.Label, val, defaultFormat, "load"))
		}
	}
	return rows
}
