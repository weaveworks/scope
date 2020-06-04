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
	if len(other) > len(m) {
		m, other = other, m
	}
	if len(other) == 0 {
		return m
	}
	result := m.Copy()
	result.UnsafeMerge(other)
	return result
}

// UnsafeMerge merges another set of Metrics into m, modifying the original.
func (m Metrics) UnsafeMerge(other Metrics) {
	if len(m) == 0 {
		if len(other) > 0 {
			m = other.Copy()
		}
		return
	}
	for k, v := range other {
		if rv, ok := m[k]; ok {
			m[k] = rv.Merge(v)
		} else {
			m[k] = v
		}
	}
}

// Copy returns a value copy of the sets map.
func (m Metrics) Copy() Metrics {
	result := make(Metrics, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Metric is a list of timeseries data with some metadata. Clients must use the
// Add method to add values.  Metrics are immutable.
type Metric struct {
	Samples []Sample `json:"samples,omitempty"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
}

func (m Metric) first() time.Time { return m.Samples[0].Timestamp }
func (m Metric) last() time.Time  { return m.Samples[len(m.Samples)-1].Timestamp }

// Sample is a single datapoint of a metric.
type Sample struct {
	Timestamp time.Time `json:"date"`
	Value     float64   `json:"value"`
}

// MakeSingletonMetric makes a metric with a single value
func MakeSingletonMetric(t time.Time, v float64) Metric {
	return Metric{
		Samples: []Sample{{t, v}},
		Min:     v,
		Max:     v,
	}

}

var emptyMetric = Metric{}

// MakeMetric makes a new Metric from unique samples incrementally ordered in
// time.
func MakeMetric(samples []Sample) Metric {
	if len(samples) < 1 {
		return emptyMetric
	}

	var (
		min = samples[0].Value
		max = samples[0].Value
	)

	for i := 1; i < len(samples); i++ {
		if samples[i].Value < min {
			min = samples[i].Value
		} else if samples[i].Value > max {
			max = samples[i].Value
		}
	}

	return Metric{
		Samples: samples,
		Min:     min,
		Max:     max,
	}
}

// WithMax returns a fresh copy of m, with Max set to max
func (m Metric) WithMax(max float64) Metric {
	return Metric{
		Samples: m.Samples,
		Max:     max,
		Min:     m.Min,
	}
}

// Len returns the number of samples in the metric.
func (m Metric) Len() int {
	return len(m.Samples)
}

// Merge combines the two Metrics and returns a new result.
func (m Metric) Merge(other Metric) Metric {

	// Optimize the empty and non-overlapping case since they are very common
	switch {
	case len(m.Samples) == 0:
		return other
	case len(other.Samples) == 0:
		return m
	case other.first().After(m.last()):
		samplesOut := make([]Sample, len(m.Samples)+len(other.Samples))
		copy(samplesOut, m.Samples)
		copy(samplesOut[len(m.Samples):], other.Samples)
		return Metric{
			Samples: samplesOut,
			Max:     math.Max(m.Max, other.Max),
			Min:     math.Min(m.Min, other.Min),
		}
	case m.first().After(other.last()):
		samplesOut := make([]Sample, len(m.Samples)+len(other.Samples))
		copy(samplesOut, other.Samples)
		copy(samplesOut[len(other.Samples):], m.Samples)
		return Metric{
			Samples: samplesOut,
			Max:     math.Max(m.Max, other.Max),
			Min:     math.Min(m.Min, other.Min),
		}
	}

	// Merge two lists of Samples in O(n)
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
	}
}

// LastSample obtains the last sample of the metric
func (m Metric) LastSample() (Sample, bool) {
	if m.Samples == nil {
		return Sample{}, false
	}
	return m.Samples[len(m.Samples)-1], true
}
