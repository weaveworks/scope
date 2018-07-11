package report_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestMetricsMerge(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)
	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)

	metrics1 := report.Metrics{
		"metric1": report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 0.1}, {Timestamp: t2, Value: 0.2}}),
		"metric2": report.MakeSingletonMetric(t3, 0.3),
	}
	metrics2 := report.Metrics{
		"metric2": report.MakeSingletonMetric(t4, 0.4),
		"metric3": report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 0.1}, {Timestamp: t2, Value: 0.2}}),
	}
	want := report.Metrics{
		"metric1": report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 0.1}, {Timestamp: t2, Value: 0.2}}),
		"metric2": report.MakeMetric([]report.Sample{{Timestamp: t3, Value: 0.3}, {Timestamp: t4, Value: 0.4}}),
		"metric3": report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 0.1}, {Timestamp: t2, Value: 0.2}}),
	}
	have := metrics1.Merge(metrics2)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func TestMetricsCopy(t *testing.T) {
	t1 := time.Now()
	want := report.Metrics{
		"metric1": report.MakeSingletonMetric(t1, 0.1),
	}
	delete(want.Copy(), "metric1") // Modify a copy
	have := want.Copy()            // Check the original wasn't affected
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func checkMetric(t *testing.T, metric report.Metric, first, last time.Time, min, max float64) {
	if !metric.First().Equal(first) {
		t.Errorf("Expected metric.First == %q, but was: %q", first, metric.First())
	}
	if !metric.Last().Equal(last) {
		t.Errorf("Expected metric.Last == %q, but was: %q", last, metric.Last())
	}
	if metric.Min != min {
		t.Errorf("Expected metric.Min == %f, but was: %f", min, metric.Min)
	}
	if metric.Max != max {
		t.Errorf("Expected metric.Max == %f, but was: %f", max, metric.Max)
	}
}

func TestMetricFirstLastMinMax(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)

	metric1 := report.MakeMetric([]report.Sample{{Timestamp: t1, Value: -0.1}, {Timestamp: t2, Value: 0.2}})

	checkMetric(t, metric1, t1, t2, -0.1, 0.2)
	checkMetric(t, metric1.Merge(metric1), t1, t2, -0.1, 0.2)

	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)
	metric2 := report.MakeMetric([]report.Sample{{Timestamp: t3, Value: 0.31}, {Timestamp: t4, Value: 0.4}})

	checkMetric(t, metric2, t3, t4, 0.31, 0.4)
	checkMetric(t, metric1.Merge(metric2), t1, t4, -0.1, 0.4)
	checkMetric(t, metric2.Merge(metric1), t1, t4, -0.1, 0.4)
}

func TestMetricMerge(t *testing.T) {
	t1 := time.Now()
	t2 := time.Now().Add(1 * time.Minute)
	t3 := time.Now().Add(2 * time.Minute)
	t4 := time.Now().Add(3 * time.Minute)

	metric1 := report.MakeMetric([]report.Sample{{Timestamp: t2, Value: 0.2}, {Timestamp: t3, Value: 0.31}})

	metric2 := report.MakeMetric([]report.Sample{{Timestamp: t1, Value: -0.1}, {Timestamp: t3, Value: 0.3}, {Timestamp: t4, Value: 0.4}})

	want := report.MakeMetric([]report.Sample{{Timestamp: t1, Value: -0.1}, {Timestamp: t2, Value: 0.2}, {Timestamp: t3, Value: 0.31}, {Timestamp: t4, Value: 0.4}})

	have := metric1.Merge(metric2)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}

	// Check it didn't modify metric1
	checkMetric(t, metric1, t2, t3, 0.2, 0.31)

	// Check the result is not the same instance as metric1
	if &metric1 == &have {
		t.Errorf("Expected different pointers for metric1 and have, but both were: %p", &have)
	}
}

func TestMetricMarshalling(t *testing.T) {
	t1 := time.Now().UTC()
	t2 := time.Now().UTC().Add(1 * time.Minute)
	t3 := time.Now().UTC().Add(2 * time.Minute)
	t4 := time.Now().UTC().Add(3 * time.Minute)

	wantSamples := []report.Sample{
		{Timestamp: t1, Value: 0.1},
		{Timestamp: t2, Value: 0.2},
		{Timestamp: t3, Value: 0.3},
		{Timestamp: t4, Value: 0.4},
	}

	want := report.MakeMetric(wantSamples)

	for _, h := range []codec.Handle{
		codec.Handle(&codec.MsgpackHandle{}),
		codec.Handle(&codec.JsonHandle{}),
	} {
		buf := &bytes.Buffer{}
		encoder := codec.NewEncoder(buf, h)
		if err := encoder.Encode(want); err != nil {
			t.Fatal(err)
		}
		bufCopy := bytes.NewBuffer(buf.Bytes())

		decoder := codec.NewDecoder(buf, h)
		var have report.Metric
		if err := decoder.Decode(&have); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}

		// extra check for samples
		decoder = codec.NewDecoder(bufCopy, h)
		var wire struct {
			Samples []report.Sample `json:"samples"`
		}
		if err := decoder.Decode(&wire); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(wantSamples, wire.Samples) {
			t.Error(test.Diff(wantSamples, wire.Samples))
		}
	}

}
