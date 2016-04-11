package report

import (
	"github.com/ugorji/go/codec"
)

// DefaultFormat and friends tell the UI how to render the "Value" of this
// metric.
const (
	DefaultFormat  = ""
	FilesizeFormat = "filesize"
	IntegerFormat  = "integer"
	PercentFormat  = "percent"
)

// MetricRow is a tuple of data used to render a metric as a sparkline and
// accoutrements.
type MetricRow struct {
	ID       string
	Label    string
	Format   string
	Group    string
	Value    float64
	Priority float64
	Metric   *Metric
}

// Summary returns a copy of the MetricRow, without the samples, just the value if there is one.
func (m MetricRow) Summary() MetricRow {
	row := m.Copy()
	if m.Metric != nil {
		row.Metric.Samples = nil
	}
	return row
}

// Copy returns a value copy of the MetricRow
func (m MetricRow) Copy() MetricRow {
	row := m
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
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Format   string   `json:"format,omitempty"`
	Group    string   `json:"group,omitempty"`
	Value    float64  `json:"value"`
	Priority float64  `json:"priority,omitempty"`
	Samples  []Sample `json:"samples"`
	Min      float64  `json:"min"`
	Max      float64  `json:"max"`
	First    string   `json:"first,omitempty"`
	Last     string   `json:"last,omitempty"`
}

// CodecEncodeSelf marshals this MetricRow. It takes the basic Metric
// rendering, then adds some row-specific fields.
func (m *MetricRow) CodecEncodeSelf(encoder *codec.Encoder) {
	in := m.Metric.ToIntermediate()
	encoder.Encode(wiredMetricRow{
		ID:       m.ID,
		Label:    m.Label,
		Format:   m.Format,
		Group:    m.Group,
		Value:    m.Value,
		Priority: m.Priority,
		Samples:  in.Samples,
		Min:      in.Min,
		Max:      in.Max,
		First:    in.First,
		Last:     in.Last,
	})
}

// CodecDecodeSelf implements codec.Selfer
func (m *MetricRow) CodecDecodeSelf(decoder *codec.Decoder) {
	var in wiredMetricRow
	decoder.Decode(&in)
	w := WireMetrics{
		Samples: in.Samples,
		Min:     in.Min,
		Max:     in.Max,
		First:   in.First,
		Last:    in.Last,
	}
	metric := w.FromIntermediate()
	*m = MetricRow{
		ID:       in.ID,
		Label:    in.Label,
		Format:   in.Format,
		Group:    in.Group,
		Value:    in.Value,
		Priority: in.Priority,
		Metric:   &metric,
	}
}

// MetricRowsByPriority implements sort.Interface, so we can sort the rows by
// priority before rendering them to the UI.
type MetricRowsByPriority []MetricRow

// Len is part of sort.Interface.
func (m MetricRowsByPriority) Len() int {
	return len(m)
}

// Swap is part of sort.Interface.
func (m MetricRowsByPriority) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Less is part of sort.Interface.
func (m MetricRowsByPriority) Less(i, j int) bool {
	return m[i].Priority < m[j].Priority
}
