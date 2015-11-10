package report

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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
// Add method to add values.
type Metric struct {
	Samples Samples   `json:"samples"`
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	First   time.Time `json:"first"`
	Last    time.Time `json:"last"`
}

// Sample is a single datapoint of a metric.
type Sample struct {
	Timestamp time.Time `json:"date"`
	Value     float64   `json:"value"`
}

// MakeMetric makes a new Metric.
func MakeMetric() Metric {
	return Metric{}
}

// WithFirst returns a fresh copy of m, with First set to t.
func (m Metric) WithFirst(t time.Time) Metric {
	m.First = t
	return m
}

// Len returns the number of samples in the metric.
func (m Metric) Len() int {
	return m.Samples.Size()
}

// Add adds the sample to the Metric. Add is the only valid way to grow a
// Metric. Add returns the Metric to enable chaining.
func (m Metric) Add(t time.Time, v float64) Metric {
	newSamples := m.Samples
	popped := Samples{}
	for !newSamples.IsNil() {
		head := newSamples.Head()
		if head.Timestamp.Equal(t) {
			// The list already has the element.
			return m
		}
		if head.Timestamp.Before(t) {
			// Reached insertion point.
			break
		}
		newSamples = newSamples.Tail()
		popped = popped.Cons(head)
	}
	newSamples = newSamples.Cons(Sample{Timestamp: t, Value: v})
	// Re-add any samples after this one.
	popped.ForEach(func(s Sample) {
		newSamples = newSamples.Cons(s)
	})
	m.Samples = newSamples
	if v > m.Max {
		m.Max = v
	}
	if v < m.Min {
		m.Min = v
	}
	if m.First.IsZero() || t.Before(m.First) {
		m.First = t
	}
	if m.Last.IsZero() || t.After(m.Last) {
		m.Last = t
	}
	return m
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {
	other.Samples.ForEach(func(s Sample) {
		m = m.Add(s.Timestamp, s.Value)
	})
	if !other.First.IsZero() && other.First.Before(m.First) {
		m.First = other.First
	}
	if !other.Last.IsZero() && other.Last.After(m.Last) {
		m.Last = other.Last
	}
	if other.Min < m.Min {
		m.Min = other.Min
	}
	if other.Max > m.Max {
		m.Max = other.Max
	}
	return m
}

// Copy returns a value copy of the Metric. Metric is immutable, so we can skip
// this.
func (m Metric) Copy() Metric {
	return m
}

// Div returns a new copy of the metric, with each value divided by n.
func (m Metric) Div(n float64) Metric {
	oldSamples := m.Samples
	m.Samples = Samples{}
	oldSamples.ForEach(func(s Sample) {
		m = m.Add(s.Timestamp, s.Value/n)
	})
	m.Max = m.Max / n
	m.Min = m.Min / n
	return m
}

// LastSample returns the last sample in the metric, or nil if there are no
// samples.
func (m Metric) LastSample() *Sample {
	if m.Samples.IsNil() {
		return nil
	}
	s := m.Samples.Head()
	return &s
}

// Samples is an immutable list of timeseries data. We have this to implement
// proper marshalling for the ps.List, as well as fixing some if ps.List's
// behaviour.
type Samples struct {
	ps.List
}

// IsNil returns true if the list is empty. Unlike ps.List, this also works if
// the list is nil.
func (s Samples) IsNil() bool {
	return s.List == nil || s.List.IsNil()
}

// Cons returns a new list with val as the head
func (s Samples) Cons(val Sample) Samples {
	if s.List == nil {
		s.List = ps.NewList()
	}
	return Samples{s.List.Cons(val)}
}

// Head returns the first element of the list;
// panics if the list is empty
func (s Samples) Head() Sample {
	return s.List.Head().(Sample)
}

// Tail returns a list with all elements except the head;
// panics if the list is empty
func (s Samples) Tail() Samples {
	return Samples{s.List.Tail()}
}

// Size returns the list's length.  This takes O(1) time. Unlike ps.List, this
// also works if the list is nil
func (s Samples) Size() int {
	if s.List == nil {
		return 0
	}
	return s.List.Size()
}

// ForEach executes a callback for each value in the list.
func (s Samples) ForEach(f func(Sample)) {
	if s.List == nil {
		return
	}
	s.List.ForEach(func(s interface{}) {
		f(s.(Sample))
	})
}

// Reverse returns a list whose elements are in the opposite order as
// the original list.
func (s Samples) Reverse() Samples {
	if s.List == nil {
		return s
	}
	return Samples{s.List.Reverse()}
}

func (s Samples) toIntermediate() []Sample {
	samples := []Sample{}
	s.Reverse().ForEach(func(s Sample) {
		samples = append(samples, s)
	})
	return samples
}

func (s Samples) fromIntermediate(in []Sample) Samples {
	list := ps.NewList()
	for _, sample := range in {
		list = list.Cons(sample)
	}
	return Samples{list}
}

// MarshalJSON implements json.Marshaller
func (s Samples) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	var err error
	if s.List == nil {
		err = json.NewEncoder(&buf).Encode(nil)
		return buf.Bytes(), err
	}

	err = json.NewEncoder(&buf).Encode(s.toIntermediate())
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Samples) UnmarshalJSON(input []byte) error {
	in := []Sample{}
	if err := json.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*s = Samples{}.fromIntermediate(in)
	return nil
}

// GobEncode implements gob.Marshaller
func (s Samples) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(s.toIntermediate())
	return buf.Bytes(), err
}

// GobDecode implements gob.Unmarshaller
func (s *Samples) GobDecode(input []byte) error {
	in := []Sample{}
	if err := gob.NewDecoder(bytes.NewBuffer(input)).Decode(&in); err != nil {
		return err
	}
	*s = Samples{}.fromIntermediate(in)
	return nil
}
