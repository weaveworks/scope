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

var (
	processNodeMetrics = renderMetrics(
		MetricRow{ID: process.CPUUsage, Label: "CPU", Format: percentFormat},
		MetricRow{ID: process.MemoryUsage, Label: "Memory", Format: filesizeFormat},
	)
	containerNodeMetrics = renderMetrics(
		MetricRow{ID: docker.CPUTotalUsage, Label: "CPU", Format: percentFormat},
		MetricRow{ID: docker.MemoryUsage, Label: "Memory", Format: filesizeFormat},
	)
	hostNodeMetrics = renderMetrics(
		MetricRow{ID: host.CPUUsage, Label: "CPU", Format: percentFormat},
		MetricRow{ID: host.MemoryUsage, Label: "Memory", Format: filesizeFormat},
		MetricRow{ID: host.Load1, Label: "Load (1m)", Format: defaultFormat, Group: "load"},
		MetricRow{ID: host.Load5, Label: "Load (5m)", Format: defaultFormat, Group: "load"},
		MetricRow{ID: host.Load15, Label: "Load (15m)", Format: defaultFormat, Group: "load"},
	)
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
	row := MetricRow{
		ID:     m.ID,
		Label:  m.Label,
		Format: m.Format,
		Group:  m.Group,
		Value:  m.Value,
	}
	if m.Metric != nil {
		var metric = m.Metric.Copy()
		row.Metric = &metric
	}
	return row
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

// NodeMetrics produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetrics(n report.Node) []MetricRow {
	renderers := map[string]func(report.Node) []MetricRow{
		report.Process:   processNodeMetrics,
		report.Container: containerNodeMetrics,
		report.Host:      hostNodeMetrics,
	}
	if renderer, ok := renderers[n.Topology]; ok {
		return renderer(n)
	}
	return nil
}

func renderMetrics(templates ...MetricRow) func(report.Node) []MetricRow {
	return func(n report.Node) []MetricRow {
		rows := []MetricRow{}
		for _, template := range templates {
			metric, ok := n.Metrics[template.ID]
			if !ok {
				continue
			}
			t := template.Copy()
			if s := metric.LastSample(); s != nil {
				t.Value = toFixed(s.Value, 2)
			}
			t.Metric = &metric
			rows = append(rows, t)
		}
		return rows
	}
}

// toFixed truncates decimals of float64 down to specified precision
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int64(num*output)) / output
}
