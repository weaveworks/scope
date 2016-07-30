package report

import (
	"math"
	"time"
)

// Metrics is a string->metric map.
type Metrics map[string]Metric

// Lookup the metric for the given key
func (m Metrics) Lookup(key string) (Metric, bool) {
	v, ok := m[key]
	return v, ok
}

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
	Samples []Sample  `json:"samples"`
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
// TODO: Specialized version adding the first sample to avoid generating garbage?
func MakeMetric() Metric {
	return Metric{}
}

// Copy returns a copy of the Metric.
func (m Metric) Copy() Metric {
	c := m
	if c.Samples != nil {
		c.Samples = make([]Sample, len(m.Samples))
		copy(c.Samples, m.Samples)
	}
	return c
}

// WithFirst returns a fresh copy of m, with first set to t.
// TODO: This seems to be unused
func (m Metric) WithFirst(t time.Time) Metric {
	return Metric{
		Samples: m.Samples,
		Max:     m.Max,
		Min:     m.Min,
		First:   t,
		Last:    m.Last,
	}
}

// WithMax returns a fresh copy of m, with Max set to max
func (m Metric) WithMax(max float64) Metric {
	return Metric{
		Samples: m.Samples,
		Max:     max,
		Min:     m.Min,
		First:   m.First,
		Last:    m.Last,
	}
}

// Len returns the number of samples in the metric.
func (m Metric) Len() int {
	return len(m.Samples)
}

func first(t1, t2 time.Time) time.Time {
	if t2.IsZero() || (!t1.IsZero() && t1.Before(t2)) {
		return t1
	}
	return t2
}

func last(t1, t2 time.Time) time.Time {
	if t2.IsZero() || (!t1.IsZero() && t1.After(t2)) {
		return t1
	}
	return t2
}

// Add returns a new Metric with (t, v) added to its Samples. Add is the only
// valid way to grow a Metric.
// TODO: join t and v into a Sample to avoid extra allocations?
// TODO: This seems to be too elaborate, Add() only seems to be used to add ordered Samples.
//       Replace this by a specialized version getting a slice of ordered Samples
//       without duplicates?
func (m Metric) Add(t time.Time, v float64) Metric {
	// Find the first element which is before you element, and insert
	// your new element in the list.  NB we want to dedupe entries with
	// equal timestamps.
	samplesOut := make([]Sample, 0, len(m.Samples)+1)
	var i int
	// TODO: use binary search + copy() to improve performance
	for i = 0; i < len(m.Samples); i++ {
		if m.Samples[i].Timestamp.Equal(t) {
			i++
			break
		}
		if m.Samples[i].Timestamp.After(t) {
			break
		}
		samplesOut = append(samplesOut, m.Samples[i])
	}
	samplesOut = append(samplesOut, Sample{t, v})
	if i < len(m.Samples) {
		samplesOut = append(samplesOut, m.Samples[i:]...)
	}
	return Metric{
		Samples: samplesOut,
		Max:     math.Max(m.Max, v),
		Min:     math.Min(m.Min, v),
		First:   first(m.First, t),
		Last:    last(m.Last, t),
	}
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {
	// Merge two lists of Samples in O(n)

	// TODO: be smarter and check for non-overlapping metrics with first and last?
	//       (copy() is much faster than checking every single sample)
	samplesOut := make([]Sample, 0, len(m.Samples)+len(other.Samples))
	mI, otherI := 0, 0
	for {
		if otherI >= len(other.Samples) {
			samplesOut = append(samplesOut, m.Samples[mI:]...)
			break
		} else if mI >= len(m.Samples) {
			samplesOut = append(samplesOut, other.Samples[otherI:]...)
			break
		}

		if m.Samples[mI].Timestamp.Equal(other.Samples[otherI].Timestamp) {
			samplesOut = append(samplesOut, m.Samples[mI])
			mI++
			otherI++
		} else if m.Samples[mI].Timestamp.Before(other.Samples[otherI].Timestamp) {
			samplesOut = append(samplesOut, m.Samples[mI])
			mI++
		} else {
			samplesOut = append(samplesOut, other.Samples[otherI])
			otherI++
		}
	}

	return Metric{
		Samples: samplesOut,
		Max:     math.Max(m.Max, other.Max),
		Min:     math.Min(m.Min, other.Min),
		First:   first(m.First, other.First),
		Last:    last(m.Last, other.Last),
	}
}

// Div returns a new copy of the metric, with each value divided by n.
func (m Metric) Div(n float64) Metric {
	samplesOut := make([]Sample, len(m.Samples), len(m.Samples))

	for i := range m.Samples {
		samplesOut[i].Value = m.Samples[i].Value / n
		samplesOut[i].Timestamp = m.Samples[i].Timestamp
	}
	return Metric{
		Samples: samplesOut,
		Max:     m.Max / n,
		Min:     m.Min / n,
		First:   m.First,
		Last:    m.Last,
	}
}

// LastSample obtains the last sample of the metric
func (m Metric) LastSample() (Sample, bool) {
	if m.Samples == nil {
		return Sample{}, false
	}
	return m.Samples[len(m.Samples)-1], true
}
