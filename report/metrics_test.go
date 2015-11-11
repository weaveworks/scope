package report_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestMetricsMerge(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)
	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)

	metrics1 := report.Metrics{
		"metric1": report.MakeMetric().Add(t1, 0.1).Add(t2, 0.2),
		"metric2": report.MakeMetric().Add(t3, 0.3),
	}
	metrics2 := report.Metrics{
		"metric2": report.MakeMetric().Add(t4, 0.4),
		"metric3": report.MakeMetric().Add(t1, 0.1).Add(t2, 0.2),
	}
	want := report.Metrics{
		"metric1": report.MakeMetric().Add(t1, 0.1).Add(t2, 0.2),
		"metric2": report.MakeMetric().Add(t3, 0.3).Add(t4, 0.4),
		"metric3": report.MakeMetric().Add(t1, 0.1).Add(t2, 0.2),
	}
	have := metrics1.Merge(metrics2)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func TestMetricsCopy(t *testing.T) {
	t1 := time.Now()
	want := report.Metrics{
		"metric1": report.MakeMetric().Add(t1, 0.1),
	}
	delete(want.Copy(), "metric1") // Modify a copy
	have := want.Copy()            // Check the original wasn't affected
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func checkMetric(t *testing.T, metric report.Metric, first, last time.Time, min, max float64) {
	if !metric.First.Equal(first) {
		t.Errorf("Expected metric.First == %q, but was: %q", first, metric.First)
	}
	if !metric.Last.Equal(last) {
		t.Errorf("Expected metric.Last == %q, but was: %q", last, metric.Last)
	}
	if metric.Min != min {
		t.Errorf("Expected metric.Min == %f, but was: %f", min, metric.Min)
	}
	if metric.Max != max {
		t.Errorf("Expected metric.Max == %f, but was: %f", max, metric.Max)
	}
}

func TestMetricFirstLastMinMax(t *testing.T) {
	metric := report.MakeMetric()
	var zero time.Time
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)
	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)
	other := report.MakeMetric()
	other.Max = 5
	other.Min = -5
	other.First = t1.Add(-1 * time.Minute)
	other.Last = t4.Add(1 * time.Minute)

	tests := []struct {
		f           func(report.Metric) report.Metric
		first, last time.Time
		min, max    float64
	}{
		{nil, zero, zero, 0, 0},
		{func(m report.Metric) report.Metric { return m.Add(t2, 2) }, t2, t2, 0, 2},
		{func(m report.Metric) report.Metric { return m.Add(t1, 1) }, t1, t2, 0, 2},
		{func(m report.Metric) report.Metric { return m.Add(t3, -1) }, t1, t3, -1, 2},
		{func(m report.Metric) report.Metric { return m.Add(t4, 3) }, t1, t4, -1, 3},
		{func(m report.Metric) report.Metric { return m.Merge(other) }, t1.Add(-1 * time.Minute), t4.Add(1 * time.Minute), -5, 5},
	}
	for _, test := range tests {
		oldFirst, oldLast, oldMin, oldMax := metric.First, metric.Last, metric.Min, metric.Max
		oldMetric := metric
		if test.f != nil {
			metric = test.f(metric)
		}

		// Check it didn't modify the old one
		checkMetric(t, oldMetric, oldFirst, oldLast, oldMin, oldMax)

		// Check the new one is as expected
		checkMetric(t, metric, test.first, test.last, test.min, test.max)
	}
}

func TestMetricAdd(t *testing.T) {
	s := []report.Sample{
		{time.Now(), 0.1},
		{time.Now().Add(1 * time.Minute), 0.2},
		{time.Now().Add(2 * time.Minute), 0.3},
	}

	intermediate := report.MakeMetric().
		Add(s[0].Timestamp, s[0].Value).
		Add(s[2].Timestamp, s[2].Value) // Keeps sorted
	have := intermediate.
		Add(s[1].Timestamp, s[1].Value).
		Add(s[2].Timestamp, 0.5) // Ignores duplicate timestamps

	intermedWant := report.Samples{}.Cons(s[0]).Cons(s[2])
	if !reflect.DeepEqual(intermedWant, intermediate.Samples) {
		t.Errorf("diff: %s", test.Diff(s, intermediate.Samples))
	}

	want := report.Samples{}.Cons(s[0]).Cons(s[1]).Cons(s[2])
	if !reflect.DeepEqual(want, have.Samples) {
		t.Errorf("diff: %s", test.Diff(want, have.Samples))
	}
}

func TestMetricMerge(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)
	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)

	metric1 := report.MakeMetric().
		Add(t2, 0.2).
		Add(t3, 0.31)

	metric2 := report.MakeMetric().
		Add(t1, -0.1).
		Add(t3, 0.3).
		Add(t4, 0.4)

	want := report.MakeMetric().
		Add(t1, -0.1).
		Add(t2, 0.2).
		Add(t3, 0.31).
		Add(t4, 0.4)
	have := metric1.Merge(metric2)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}

	// Check it didn't modify metric1
	if !metric1.First.Equal(t2) {
		t.Errorf("Expected metric1.First == %q, but was: %q", t2, metric1.First)
	}
	if !metric1.Last.Equal(t3) {
		t.Errorf("Expected metric1.Last == %q, but was: %q", t3, metric1.Last)
	}
	if metric1.Min != 0.0 {
		t.Errorf("Expected metric1.Min == %f, but was: %f", 0.0, metric1.Min)
	}
	if metric1.Max != 0.31 {
		t.Errorf("Expected metric1.Max == %f, but was: %f", 0.31, metric1.Max)
	}

	// Check the result is not the same instance as metric1
	if &metric1 == &have {
		t.Errorf("Expected different pointers for metric1 and have, but both were: %p", &have)
	}
}

func TestMetricCopy(t *testing.T) {
	want := report.MakeMetric()
	have := want.Copy()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}

	want = report.MakeMetric().Add(time.Now(), 1)
	have = want.Copy()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func TestMetricDiv(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)

	want := report.MakeMetric().
		Add(t1, -2).
		Add(t2, 2)
	beforeDiv := report.MakeMetric().
		Add(t1, -2048).
		Add(t2, 2048)
	have := beforeDiv.Div(1024)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}

	// Check the original was unmodified
	checkMetric(t, beforeDiv, t1, t2, -2048, 2048)
}

func TestMetricMarshalling(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC().Add(1 * time.Minute)
	t3 := time.Now().UTC().Add(2 * time.Minute)
	t4 := time.Now().UTC().Add(3 * time.Minute)

	want := report.MakeMetric().
		Add(t1, 0.1).
		Add(t2, 0.2).
		Add(t3, 0.3).
		Add(t4, 0.4)

	codecs := []struct {
		name   string
		encode func(interface{}) ([]byte, error)
		decode func([]byte, interface{}) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{
			"gob",
			func(v interface{}) ([]byte, error) {
				buf := &bytes.Buffer{}
				err := gob.NewEncoder(buf).Encode(v)
				return buf.Bytes(), err
			},
			func(b []byte, v interface{}) error {
				return gob.NewDecoder(bytes.NewReader(b)).Decode(v)
			},
		},
	}
	for _, codec := range codecs {
		b, err := codec.encode(want)
		if err != nil {
			t.Fatalf("[%s] %s", codec.name, err)
		}

		var have report.Metric
		err = codec.decode(b, &have)
		if err != nil {
			t.Fatalf("[%s] %s", codec.name, err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Errorf("[%s] diff: %sencoded: %s", codec.name, test.Diff(want, have), b)
		}
	}
}
