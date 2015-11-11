package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"math"
	"time"

	"github.com/mndrix/ps"
)

// Metrics is a string->metric map.
type Metrics map[string]Metric

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (m Metrics) Merge(other Metrics) Metrics {
	result := m.Copy()
	for k, v := range other {
		result[k] = result[k].Merge(v)
	}
	return result
}

// Copy returns a value copy of the sets map.
func (m Metrics) Copy() Metrics {
	result := Metrics{}
	for k, v := range m {
		result[k] = v.Copy()
	}
	return result
}

// Metric is a list of timeseries data with some metadata. Clients must use the
// Add method to add values.  Metrics are immutable.
type Metric struct {
	Samples     ps.List
	Min, Max    float64
	First, Last time.Time
}

// Sample is a single datapoint of a metric.
type Sample struct {
	Timestamp time.Time `json:"date"`
	Value     float64   `json:"value"`
}

// MakeMetric makes a new Metric.
func MakeMetric() Metric {
	return Metric{
		Samples: ps.NewList(),
	}
}

// Copy returns a value copy of the Metric. Metric is immutable, so we can skip
// this.
func (m Metric) Copy() Metric {
	return m
}

// WithFirst returns a fresh copy of m, with First set to t.
func (m Metric) WithFirst(t time.Time) Metric {
	return Metric{
		Samples: m.Samples,
		Max:     m.Max,
		Min:     m.Min,
		First:   t,
		Last:    m.Last,
	}
}

// Len returns the number of samples in the metric.
func (m Metric) Len() int {
	if m.Samples == nil {
		return 0
	}
	return m.Samples.Size()
}

func first(t1, t2 time.Time) time.Time {
	if !t1.IsZero() && t1.Before(t2) {
		return t1
	}
	return t2
}

func last(t1, t2 time.Time) time.Time {
	if !t1.IsZero() && t1.After(t2) {
		return t1
	}
	return t2
}

// Add returns a new Metric with (t, v) added to its Samples. Add is the only
// valid way to grow a Metric.
func (m Metric) Add(t time.Time, v float64) Metric {
	// Find the first element which is before you element, and insert
	// your new element in the list.  NB we want to dedupe entries with
	// equal timestamps.
	var insert func(ps.List) ps.List
	insert = func(ss ps.List) ps.List {
		if ss == nil || ss.IsNil() {
			return ps.NewList().Cons(Sample{t, v})
		}
		currSample := ss.Head().(Sample)
		if currSample.Timestamp.Equal(t) {
			return ss.Tail().Cons(Sample{t, v})
		}
		if currSample.Timestamp.Before(t) {
			return ss.Cons(Sample{t, v})
		}
		return insert(ss.Tail()).Cons(currSample)
	}

	return Metric{
		Samples: insert(m.Samples),
		Max:     math.Max(m.Max, v),
		Min:     math.Min(m.Min, v),
		First:   first(m.First, t),
		Last:    last(m.Last, t),
	}
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {
	var merge func(ps.List, ps.List) ps.List
	merge = func(ss1, ss2 ps.List) ps.List {
		if ss1 == nil || ss1.IsNil() {
			return ss2
		} else if ss2 == nil || ss2.IsNil() {
			return ss1
		}

		s1 := ss1.Head().(Sample)
		s2 := ss2.Head().(Sample)

		if s1.Timestamp.Equal(s2.Timestamp) {
			return merge(ss1.Tail(), ss2.Tail()).Cons(s1)
		} else if s1.Timestamp.After(s2.Timestamp) {
			return merge(ss1.Tail(), ss2).Cons(s1)
		} else {
			return merge(ss1, ss2.Tail()).Cons(s2)
		}
	}

	return Metric{
		Samples: merge(m.Samples, other.Samples),
		Max:     math.Max(m.Max, other.Max),
		Min:     math.Min(m.Min, other.Min),
		First:   first(m.First, other.First),
		Last:    last(m.Last, other.Last),
	}
}

// Div returns a new copy of the metric, with each value divided by n.
func (m Metric) Div(n float64) Metric {
	var div func(ps.List) ps.List
	div = func(ss ps.List) ps.List {
		if ss == nil || ss.IsNil() {
			return ss
		}
		s := ss.Head().(Sample)
		return div(ss.Tail()).Cons(Sample{s.Timestamp, s.Value / n})
	}
	return Metric{
		Samples: div(m.Samples),
		Max:     m.Max / n,
		Min:     m.Min / n,
		First:   m.First,
		Last:    m.Last,
	}
}

// LastSample returns the last sample in the metric, or nil if there are no
// samples.
func (m Metric) LastSample() *Sample {
	if m.Samples == nil || m.Samples.IsNil() {
		return nil
	}
	s := m.Samples.Head().(Sample)
	return &s
}

// WireMetrics is the on-the-wire representation of Metrics.
type WireMetrics struct {
	Samples []Sample  `json:"samples"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	First   time.Time `json:"first"`
	Last    time.Time `json:"last"`
}

func (m Metric) toIntermediate() WireMetrics {
	samples := []Sample{}
	if m.Samples != nil {
		m.Samples.ForEach(func(s interface{}) {
			samples = append(samples, s.(Sample))
		})
	}
	return WireMetrics{
		Samples: samples,
		Max:     m.Max,
		Min:     m.Min,
		First:   m.First,
		Last:    m.Last,
	}
}

func (m WireMetrics) fromIntermediate() Metric {
	samples := ps.NewList()
	for _, s := range m.Samples {
		samples = samples.Cons(s)
	}
	return Metric{
		Samples: samples.Reverse(),
		Max:     m.Max,
		Min:     m.Min,
		First:   m.First,
		Last:    m.Last,
	}
}

// MarshalJSON implements json.Marshaller
func (m Metric) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	in := m.toIntermediate()
	err := json.NewEncoder(&buf).Encode(in)
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Metric) UnmarshalJSON(input []byte) error {
	in := WireMetrics{}
	if err := json.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*m = in.fromIntermediate()
	return nil
}

// GobEncode implements gob.Marshaller
func (m Metric) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(m.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (m *Metric) GobDecode(input []byte) error {
	in := WireMetrics{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*m = in.fromIntermediate()
	return nil
}
