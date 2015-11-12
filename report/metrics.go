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
		result[k] = v
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

var nilMetric = Metric{Samples: ps.NewList()}

// MakeMetric makes a new Metric.
func MakeMetric() Metric {
	return nilMetric
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

// revCons appends acc to the head of curr, where acc is in reverse order.
// acc must never be nil, curr can be.
func revCons(acc, curr ps.List) ps.List {
	if curr == nil {
		return acc.Reverse()
	}
	for !acc.IsNil() {
		acc, curr = acc.Tail(), curr.Cons(acc.Head())
	}
	return curr
}

// Add returns a new Metric with (t, v) added to its Samples. Add is the only
// valid way to grow a Metric.
func (m Metric) Add(t time.Time, v float64) Metric {
	// Find the first element which is before you element, and insert
	// your new element in the list.  NB we want to dedupe entries with
	// equal timestamps.
	// This should be O(1) to insert a latest element, and O(n) in general.
	curr, acc := m.Samples, ps.NewList()
	for {
		if curr == nil || curr.IsNil() {
			acc = acc.Cons(Sample{t, v})
			break
		}

		currSample := curr.Head().(Sample)
		if currSample.Timestamp.Equal(t) {
			acc, curr = acc.Cons(Sample{t, v}), curr.Tail()
			break
		}
		if currSample.Timestamp.Before(t) {
			acc = acc.Cons(Sample{t, v})
			break
		}

		acc, curr = acc.Cons(curr.Head()), curr.Tail()
	}
	acc = revCons(acc, curr)

	return Metric{
		Samples: acc,
		Max:     math.Max(m.Max, v),
		Min:     math.Min(m.Min, v),
		First:   first(m.First, t),
		Last:    last(m.Last, t),
	}
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {
	// Merge two lists of samples in O(n)
	curr1, curr2, acc := m.Samples, other.Samples, ps.NewList()

	for {
		if curr1 == nil || curr1.IsNil() {
			acc = revCons(acc, curr2)
			break
		} else if curr2 == nil || curr2.IsNil() {
			acc = revCons(acc, curr1)
			break
		}

		s1 := curr1.Head().(Sample)
		s2 := curr2.Head().(Sample)

		if s1.Timestamp.Equal(s2.Timestamp) {
			curr1, curr2, acc = curr1.Tail(), curr2.Tail(), acc.Cons(s1)
		} else if s1.Timestamp.After(s2.Timestamp) {
			curr1, acc = curr1.Tail(), acc.Cons(s1)
		} else {
			curr2, acc = curr2.Tail(), acc.Cons(s2)
		}
	}

	return Metric{
		Samples: acc,
		Max:     math.Max(m.Max, other.Max),
		Min:     math.Min(m.Min, other.Min),
		First:   first(m.First, other.First),
		Last:    last(m.Last, other.Last),
	}
}

// Div returns a new copy of the metric, with each value divided by n.
func (m Metric) Div(n float64) Metric {
	curr, acc := m.Samples, ps.NewList()
	for curr != nil && !curr.IsNil() {
		s := curr.Head().(Sample)
		curr, acc = curr.Tail(), acc.Cons(Sample{s.Timestamp, s.Value / n})
	}
	acc = acc.Reverse()
	return Metric{
		Samples: acc,
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
	Samples []Sample  `json:"samples"` // On the wire, samples are sorted oldest to newest,
	Min     float64   `json:"min"`     // the opposite order to how we store them internally.
	Max     float64   `json:"max"`
	First   time.Time `json:"first"`
	Last    time.Time `json:"last"`
}

func (m Metric) toIntermediate() WireMetrics {
	samples := []Sample{}
	if m.Samples != nil {
		m.Samples.Reverse().ForEach(func(s interface{}) {
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
		Samples: samples,
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
