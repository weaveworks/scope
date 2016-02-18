package detailed

import (
	"math"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

const (
	defaultFormat  = ""
	filesizeFormat = "filesize"
	integerFormat  = "integer"
	percentFormat  = "percent"
)

var (
	processNodeMetrics = []MetricRow{
		{ID: process.CPUUsage, Format: percentFormat},
		{ID: process.MemoryUsage, Format: filesizeFormat},
		{ID: process.OpenFilesCount, Format: integerFormat},
	}
	containerNodeMetrics = []MetricRow{
		{ID: docker.CPUTotalUsage, Format: percentFormat},
		{ID: docker.MemoryUsage, Format: filesizeFormat},
	}
	hostNodeMetrics = []MetricRow{
		{ID: host.CPUUsage, Format: percentFormat},
		{ID: host.MemoryUsage, Format: filesizeFormat},
		{ID: host.Load1, Format: defaultFormat, Group: "load"},
		{ID: host.Load5, Format: defaultFormat, Group: "load"},
		{ID: host.Load15, Format: defaultFormat, Group: "load"},
	}
)

// MetricRow is a tuple of data used to render a metric as a sparkline and
// accoutrements.
type MetricRow struct {
	ID     string
	Format string
	Group  string
	Value  float64
	Metric *report.Metric
}

// Copy returns a value copy of the MetricRow
func (m MetricRow) Copy() MetricRow {
	row := MetricRow{
		ID:     m.ID,
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

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (MetricRow) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*MetricRow) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}

type wiredMetricRow struct {
	ID      string          `json:"id"`
	Label   string          `json:"label"`
	Format  string          `json:"format,omitempty"`
	Group   string          `json:"group,omitempty"`
	Value   float64         `json:"value"`
	Samples []report.Sample `json:"samples"`
	Min     float64         `json:"min"`
	Max     float64         `json:"max"`
	First   string          `json:"first,omitempty"`
	Last    string          `json:"last,omitempty"`
}

// CodecEncodeSelf marshals this MetricRow. It takes the basic Metric
// rendering, then adds some row-specific fields.
func (m *MetricRow) CodecEncodeSelf(encoder *codec.Encoder) {
	in := m.Metric.ToIntermediate()
	encoder.Encode(wiredMetricRow{
		ID:      m.ID,
		Label:   Label(m.ID),
		Format:  m.Format,
		Group:   m.Group,
		Value:   m.Value,
		Samples: in.Samples,
		Min:     in.Min,
		Max:     in.Max,
		First:   in.First,
		Last:    in.Last,
	})
}

// CodecDecodeSelf implements codec.Selfer
func (m *MetricRow) CodecDecodeSelf(decoder *codec.Decoder) {
	var in wiredMetricRow
	decoder.Decode(&in)
	w := report.WireMetrics{
		Samples: in.Samples,
		Min:     in.Min,
		Max:     in.Max,
		First:   in.First,
		Last:    in.Last,
	}
	metric := w.FromIntermediate()
	*m = MetricRow{
		ID:     in.ID,
		Format: in.Format,
		Group:  in.Group,
		Value:  in.Value,
		Metric: &metric,
	}
}

// NodeMetrics produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func NodeMetrics(n report.Node) []MetricRow {
	renderers := map[string][]MetricRow{
		report.Process:   processNodeMetrics,
		report.Container: containerNodeMetrics,
		report.Host:      hostNodeMetrics,
	}
	if templates, ok := renderers[n.Topology]; ok {
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
	return nil
}

// toFixed truncates decimals of float64 down to specified precision
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int64(num*output)) / output
}
