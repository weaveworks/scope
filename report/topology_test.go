package report_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
)

func TestMakeStringSet(t *testing.T) {
	for _, testcase := range []struct {
		input []string
		want  report.StringSet
	}{
		{input: nil, want: nil},
		{input: []string{}, want: report.MakeStringSet()},
		{input: []string{"a"}, want: report.MakeStringSet("a")},
		{input: []string{"a", "a"}, want: report.MakeStringSet("a")},
		{input: []string{"b", "c", "a"}, want: report.MakeStringSet("a", "b", "c")},
	} {
		if want, have := testcase.want, report.MakeStringSet(testcase.input...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v: want %v, have %v", testcase.input, want, have)
		}
	}
}

func TestStringSetAdd(t *testing.T) {
	for _, testcase := range []struct {
		input report.StringSet
		strs  []string
		want  report.StringSet
	}{
		{input: report.StringSet(nil), strs: []string{}, want: report.StringSet(nil)},
		{input: report.MakeStringSet(), strs: []string{}, want: report.MakeStringSet()},
		{input: report.MakeStringSet("a"), strs: []string{}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet(), strs: []string{"a"}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a"), strs: []string{"a"}, want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("b"), strs: []string{"a", "b"}, want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("a"), strs: []string{"c", "b"}, want: report.MakeStringSet("a", "b", "c")},
		{input: report.MakeStringSet("a", "c"), strs: []string{"b", "b", "b"}, want: report.MakeStringSet("a", "b", "c")},
	} {
		if want, have := testcase.want, testcase.input.Add(testcase.strs...); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.strs, want, have)
		}
	}
}

func TestStringSetMerge(t *testing.T) {
	for _, testcase := range []struct {
		input report.StringSet
		other report.StringSet
		want  report.StringSet
	}{
		{input: report.StringSet(nil), other: report.StringSet(nil), want: report.StringSet(nil)},
		{input: report.MakeStringSet(), other: report.MakeStringSet(), want: report.MakeStringSet()},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet(), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet(), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("b"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
		{input: report.MakeStringSet("a"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a")},
		{input: report.MakeStringSet("a", "c"), other: report.MakeStringSet("a", "b"), want: report.MakeStringSet("a", "b", "c")},
		{input: report.MakeStringSet("b"), other: report.MakeStringSet("a"), want: report.MakeStringSet("a", "b")},
	} {
		if want, have := testcase.want, testcase.input.Merge(testcase.other); !reflect.DeepEqual(want, have) {
			t.Errorf("%v + %v: want %v, have %v", testcase.input, testcase.other, want, have)
		}
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
		if test.f != nil {
			metric = test.f(metric)
		}

		if !metric.First.Equal(test.first) {
			t.Errorf("Expected metric.First == %q, but was: %q", test.last, metric.First)
		}
		if !metric.Last.Equal(test.last) {
			t.Errorf("Expected metric.Last == %q, but was: %q", test.first, metric.Last)
		}
		if metric.Min != test.min {
			t.Errorf("Expected metric.Min == %f, but was: %f", test.min, metric.Min)
		}
		if metric.Max != test.max {
			t.Errorf("Expected metric.Max == %f, but was: %f", test.max, metric.Max)
		}
	}
}

func TestMetricAdd(t *testing.T) {
	want := []report.Sample{
		{time.Now(), 0.1},
		{time.Now().Add(1 * time.Minute), 0.2},
		{time.Now().Add(2 * time.Minute), 0.3},
	}
	have := report.MakeMetric().
		Add(want[0].Timestamp, want[0].Value).
		Add(want[2].Timestamp, want[2].Value). // Keeps sorted
		Add(want[1].Timestamp, want[1].Value).
		Add(want[2].Timestamp, 0.5) // Ignores duplicate timestamps

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
		Add(t1, 0.1).
		Add(t3, 0.31).
		Add(t4, 0.4)

	metric2 := report.MakeMetric().
		Add(t2, 0.2).
		Add(t3, 0.3)

	want := report.MakeMetric().
		Add(t1, 0.1).
		Add(t2, 0.2).
		Add(t3, 0.31).
		Add(t4, 0.4)
	have := metric1.Merge(metric2)
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}

func TestMetricCopy(t *testing.T) {
	var want, have report.Metric
	want = report.MakeMetric()
	have = want.Copy()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}

	want = report.MakeMetric().Add(time.Now(), 1)
	have = want.Copy()
	if !reflect.DeepEqual(want, have) {
		t.Errorf("diff: %s", test.Diff(want, have))
	}
}
